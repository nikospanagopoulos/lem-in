package main

import (
	"errors"
	"fmt"
	"math"
	"unicode/utf8"

	"lem-in/internal/graph"
)

// errTooSmall is returned when the drawing area cannot hold the rooms.
var errTooSmall = errors.New("the drawing area is too small")

// The coordinates of the rooms in the input file can be any integers: 0 to 23
// in one colony, 0 to 5000 in another. The terminal is a grid of characters,
// about 80 columns and 24 rows. The layout turns the first into the second:
// it decides, for every room, in which column and row of the grid it is drawn.
//
// A room is drawn as "[name]", so a room is not one cell but a short stretch
// of a row. Its position is the column of the "[".

// position is a cell of the grid.
type position struct {
	Col, Row int
}

// layout holds the grid size and the position of every room.
type layout struct {
	width, height int
	rooms         map[string]position
	// labelWidth[name] is the number of columns the label of a room needs.
	labelWidth map[string]int
}

// label returns how the room is written on the grid: "[name]".
func label(name string) string { return "[" + name + "]" }

// newLayout places the rooms on a grid of width columns and height rows.
//
// The rule is a linear stretch. The smallest x of all rooms goes to the left
// edge, the biggest x to the right edge, and the others in proportion. The
// same is done for y. So the colony always fills the area, and its shape is
// kept: a room that is left of another in the file is left of it on the grid.
// The y axis grows downward, like the rows of a terminal.
func newLayout(rooms []graph.Room, width, height int) (layout, error) {
	l := layout{
		width:      width,
		height:     height,
		rooms:      make(map[string]position, len(rooms)),
		labelWidth: make(map[string]int, len(rooms)),
	}
	if len(rooms) == 0 {
		return l, nil
	}
	if height < 1 {
		return layout{}, fmt.Errorf("%w: height %d", errTooSmall, height)
	}

	// The widest label limits how far right a room can start, so that no label
	// goes outside the grid.
	widest := 0
	minX, maxX := rooms[0].X, rooms[0].X
	minY, maxY := rooms[0].Y, rooms[0].Y
	for _, room := range rooms {
		w := utf8.RuneCountInString(label(room.Name))
		l.labelWidth[room.Name] = w
		widest = max(widest, w)
		minX, maxX = min(minX, room.X), max(maxX, room.X)
		minY, maxY = min(minY, room.Y), max(maxY, room.Y)
	}
	if width < widest {
		return layout{}, fmt.Errorf("%w: width %d, but a room needs %d", errTooSmall, width, widest)
	}

	// Columns a label can start in: 0 .. lastCol. Rows: 0 .. lastRow.
	lastCol, lastRow := width-widest, height-1

	taken := make([][]span, height) // for each row, the stretches already used
	for _, room := range rooms {
		col := stretch(room.X, minX, maxX, lastCol)
		row := stretch(room.Y, minY, maxY, lastRow)
		w := l.labelWidth[room.Name]

		// Two close rooms can land on the same cell, or labels can overlap.
		// Then the room moves to the nearest row where it fits. Rooms are
		// handled in file order, so the result is the same on every run.
		row = freeRow(taken, row, span{col, col + w})
		taken[row] = append(taken[row], span{col, col + w})
		l.rooms[room.Name] = position{Col: col, Row: row}
	}
	return l, nil
}

// stretch maps v from the range [lo, hi] to the range [0, last].
func stretch(v, lo, hi, last int) int {
	if hi == lo {
		return last / 2 // all values are equal: put it in the middle
	}
	// float64 avoids an overflow of the multiplication for huge coordinates.
	fraction := float64(v-lo) / float64(hi-lo)
	return int(math.Round(fraction * float64(last)))
}

// span is a stretch of columns [from, to) of one row.
type span struct{ from, to int }

// overlaps reports whether two stretches touch or cross. One free column is
// kept between two labels, so "[a][b]" is never drawn as one block.
func (s span) overlaps(o span) bool {
	return s.from < o.to+1 && o.from < s.to+1
}

// freeRow returns the row nearest to want, where the stretch fits without
// overlapping another one. It tries want, then want-1, want+1, want-2, ... If
// the area is so full that no row is free, the room stays in its row.
func freeRow(taken [][]span, want int, s span) int {
	for distance := 0; distance < len(taken); distance++ {
		for _, row := range []int{want - distance, want + distance} {
			if row >= 0 && row < len(taken) && fits(taken[row], s) {
				return row
			}
		}
	}
	return want
}

func fits(used []span, s span) bool {
	for _, other := range used {
		if s.overlaps(other) {
			return false
		}
	}
	return true
}

// at returns the position of a room, and false if the room is unknown.
func (l layout) at(name string) (position, bool) {
	p, ok := l.rooms[name]
	return p, ok
}

// centre returns the cell in the middle of a room's label. The tunnels are
// drawn from centre to centre.
func (l layout) centre(name string) (position, bool) {
	p, ok := l.rooms[name]
	if !ok {
		return position{}, false
	}
	return position{Col: p.Col + l.labelWidth[name]/2, Row: p.Row}, true
}
