package main

import (
	"fmt"
	"strings"

	"lem-in/internal/graph"
)

// This file draws one frame: the rooms, the tunnels between them, and where
// the ants are. It returns text and prints nothing, so it is easy to test.
//
// A frame has three layers, drawn in this order:
//
//  1. the tunnels, as lines of characters from the centre of one room to the
//     centre of the other;
//  2. the rooms, as "[name]", written over the tunnels so a tunnel never cuts
//     through a name;
//  3. below the grid, a few lines of text with the ants.

// style says how a cell is coloured when colours are on.
type style int

const (
	styleNone   style = iota // empty cell, or plain text
	styleTunnel              // part of a tunnel (not coloured, but remembered)
	styleStart               // the ##start room
	styleEnd                 // the ##end room
	styleAnts                // a room that holds ants right now
)

// ANSI escape codes. "\x1b[" starts a code, the number picks the effect and
// "m" ends it. 0 resets everything.
const (
	ansiReset = "\x1b[0m"
	ansiStart = "\x1b[32m"   // green
	ansiEnd   = "\x1b[31m"   // red
	ansiAnts  = "\x1b[1;33m" // bold yellow
)

// cell is one character of the grid with its style.
type cell struct {
	r  rune
	st style
}

// canvas is the grid of characters a frame is drawn on.
type canvas struct {
	cells [][]cell
}

func newCanvas(width, height int) *canvas {
	cells := make([][]cell, height)
	for row := range cells {
		cells[row] = make([]cell, width)
		for col := range cells[row] {
			cells[row][col] = cell{r: ' '}
		}
	}
	return &canvas{cells: cells}
}

// inside reports whether a cell exists. Drawing outside the grid does nothing,
// so the drawing code never needs to worry about the edges.
func (c *canvas) inside(col, row int) bool {
	return row >= 0 && row < len(c.cells) && col >= 0 && col < len(c.cells[row])
}

// drawLabel writes text starting at p, all in one style.
func (c *canvas) drawLabel(p position, text string, st style) {
	col := p.Col
	for _, r := range text {
		if c.inside(col, p.Row) {
			c.cells[p.Row][col] = cell{r: r, st: st}
		}
		col++
	}
}

// putTunnel puts one piece of tunnel in a cell. Where two tunnels with a
// different direction meet, the cell becomes '+', so a crossing is visible.
func (c *canvas) putTunnel(col, row int, glyph rune) {
	if !c.inside(col, row) {
		return
	}
	old := c.cells[row][col]
	if old.st == styleTunnel && old.r != glyph {
		glyph = '+'
	}
	c.cells[row][col] = cell{r: glyph, st: styleTunnel}
}

// drawTunnel draws a line of characters from cell a to cell b.
//
// It uses Bresenham's algorithm: it walks from a to b one cell at a time and
// each time decides whether to step sideways, up/down, or both, so that the
// cells stay as close as possible to the real straight line. It needs only
// integers.
func (c *canvas) drawTunnel(a, b position) {
	glyph := tunnelGlyph(a, b)

	dx, dy := abs(b.Col-a.Col), abs(b.Row-a.Row)
	stepCol, stepRow := sign(b.Col-a.Col), sign(b.Row-a.Row)
	// err is how far the cells so far are from the true line.
	err := dx - dy

	col, row := a.Col, a.Row
	for {
		c.putTunnel(col, row, glyph)
		if col == b.Col && row == b.Row {
			return
		}
		twice := 2 * err
		if twice > -dy {
			err -= dy
			col += stepCol
		}
		if twice < dx {
			err += dx
			row += stepRow
		}
	}
}

// tunnelGlyph chooses the character for a tunnel from its direction.
//
// A terminal cell is about twice as tall as it is wide. So a line that goes
// 2 columns right for each row down looks like a 45 degree line, and that is
// what '╲' and '╱' show. A flatter line is drawn with '─' and a steeper one
// with '│'.
func tunnelGlyph(a, b position) rune {
	dCol, dRow := b.Col-a.Col, b.Row-a.Row
	switch {
	case dRow == 0:
		return '─'
	case dCol == 0:
		return '│'
	}
	ratio := float64(abs(dCol)) / float64(2*abs(dRow))
	switch {
	case ratio >= 2:
		return '─'
	case ratio <= 0.5:
		return '│'
	case (dCol > 0) == (dRow > 0):
		return '╲' // going right and down, or left and up
	default:
		return '╱' // going right and up, or left and down
	}
}

// text returns the grid as text, one line per row. Spaces at the end of a line
// are cut. With colour on, each style change becomes an ANSI code.
func (c *canvas) text(colour bool) string {
	var b strings.Builder
	for _, row := range c.cells {
		end := len(row)
		for end > 0 && row[end-1].r == ' ' && row[end-1].st == styleNone {
			end--
		}

		current := styleNone
		for _, cl := range row[:end] {
			if colour && cl.st != current {
				if current != styleNone {
					b.WriteString(ansiReset)
				}
				b.WriteString(code(cl.st))
				current = cl.st
			}
			b.WriteRune(cl.r)
		}
		if current != styleNone {
			b.WriteString(ansiReset)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// code returns the ANSI code for a style, or "" for a style with no colour.
func code(st style) string {
	switch st {
	case styleStart:
		return ansiStart
	case styleEnd:
		return ansiEnd
	case styleAnts:
		return ansiAnts
	}
	return ""
}

// renderFrame draws the colony with the ants of one moment. occupancy tells
// which ants are in which room (see state.occupancy).
//
// The result is the grid, then two lines of text:
//
//	start [1]: 2 ants   end [0]: 0 ants
//	ants in rooms: [3] L1  [2] L2
func renderFrame(colony graph.Colony, l layout, occupancy map[string][]int, colour bool) string {
	c := newCanvas(l.width, l.height)

	// Layer 1: tunnels.
	for _, link := range colony.Links {
		from, okFrom := l.centre(link[0])
		to, okTo := l.centre(link[1])
		if okFrom && okTo {
			c.drawTunnel(from, to)
		}
	}

	// Layer 2: rooms, over the tunnels.
	for _, room := range colony.Rooms {
		p, ok := l.at(room.Name)
		if !ok {
			continue
		}
		st := styleNone
		switch {
		case room.Name == colony.Start:
			st = styleStart
		case room.Name == colony.End:
			st = styleEnd
		case len(occupancy[room.Name]) > 0:
			st = styleAnts
		}
		c.drawLabel(p, label(room.Name), st)
	}

	// Layer 3: the text below the grid.
	var b strings.Builder
	b.WriteString(c.text(colour))
	fmt.Fprintf(&b, "start [%s]: %s   end [%s]: %s\n",
		colony.Start, plural(len(occupancy[colony.Start])),
		colony.End, plural(len(occupancy[colony.End])))
	b.WriteString(antsInRooms(colony, occupancy))
	b.WriteByte('\n')
	return b.String()
}

// antsInRooms lists the ants that are in the rooms between start and end, in
// the order of the rooms in the file, so the line is the same on every run.
func antsInRooms(colony graph.Colony, occupancy map[string][]int) string {
	var parts []string
	for _, room := range colony.Rooms {
		if room.Name == colony.Start || room.Name == colony.End {
			continue
		}
		ants := occupancy[room.Name]
		if len(ants) == 0 {
			continue
		}
		var b strings.Builder
		b.WriteString(label(room.Name))
		for _, ant := range ants {
			fmt.Fprintf(&b, " L%d", ant)
		}
		parts = append(parts, b.String())
	}
	if len(parts) == 0 {
		return "ants in rooms: none"
	}
	return "ants in rooms: " + strings.Join(parts, "  ")
}

// plural writes "1 ant" or "3 ants".
func plural(n int) string {
	if n == 1 {
		return "1 ant"
	}
	return fmt.Sprintf("%d ants", n)
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	}
	return 0
}
