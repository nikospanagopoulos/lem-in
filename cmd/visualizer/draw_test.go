package main

import (
	"path/filepath"
	"strings"
	"testing"

	"lem-in/internal/graph"
	"lem-in/internal/parser"
)

// twoRooms is the smallest colony with a tunnel: s -- e, on one row.
func twoRooms() graph.Colony {
	return graph.Colony{
		Ants:  1,
		Start: "s",
		End:   "e",
		Rooms: []graph.Room{{Name: "s", X: 0, Y: 0}, {Name: "e", X: 1, Y: 0}},
		Links: [][2]string{{"s", "e"}},
	}
}

func frameLines(t *testing.T, colony graph.Colony, width, height int, occ map[string][]int, colour bool) []string {
	t.Helper()
	l, err := newLayout(colony.Rooms, width, height)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(renderFrame(colony, l, occ, colour), "\n")
}

// A horizontal tunnel: "[s]", 14 dashes, "[e]". The tunnel runs from the
// centre of one label to the centre of the other and the labels are written
// over its ends.
func TestRenderFrameHorizontalTunnel(t *testing.T) {
	lines := frameLines(t, twoRooms(), 20, 3, map[string][]int{"s": {1}}, false)

	if want := "[s]" + strings.Repeat("─", 14) + "[e]"; lines[1] != want {
		t.Errorf("row 1 = %q, want %q", lines[1], want)
	}
	if lines[0] != "" || lines[2] != "" {
		t.Errorf("the other rows are not empty: %q %q", lines[0], lines[2])
	}
	if want := "start [s]: 1 ant   end [e]: 0 ants"; lines[3] != want {
		t.Errorf("status line = %q, want %q", lines[3], want)
	}
	if want := "ants in rooms: none"; lines[4] != want {
		t.Errorf("ants line = %q, want %q", lines[4], want)
	}
}

// A vertical tunnel: rooms a and b one above the other.
func TestRenderFrameVerticalTunnel(t *testing.T) {
	colony := graph.Colony{
		Ants: 1, Start: "a", End: "b",
		Rooms: []graph.Room{{Name: "a", X: 0, Y: 0}, {Name: "b", X: 0, Y: 9}},
		Links: [][2]string{{"a", "b"}},
	}
	lines := frameLines(t, colony, 21, 6, nil, false)

	// "[a]" starts in column 9, so its centre is column 10.
	for row := 1; row <= 4; row++ {
		if want := strings.Repeat(" ", 10) + "│"; lines[row] != want {
			t.Errorf("row %d = %q, want %q", row, lines[row], want)
		}
	}
}

func TestTunnelGlyph(t *testing.T) {
	tests := []struct {
		name string
		a, b position
		want rune
	}{
		{"horizontal", position{0, 0}, position{10, 0}, '─'},
		{"vertical", position{0, 0}, position{0, 5}, '│'},
		{"right and down", position{0, 0}, position{10, 5}, '╲'},
		{"left and up", position{10, 5}, position{0, 0}, '╲'},
		{"right and up", position{0, 5}, position{10, 0}, '╱'},
		{"left and down", position{10, 0}, position{0, 5}, '╱'},
		{"very flat", position{0, 0}, position{40, 5}, '─'},
		{"very steep", position{0, 0}, position{2, 5}, '│'},
		// The limits: a line exactly 2 columns per row is still flat, and
		// exactly half a column per row is still steep.
		{"flat limit", position{0, 0}, position{20, 5}, '─'},
		{"steep limit", position{0, 0}, position{5, 5}, '│'},
		{"just inside the diagonal, flat side", position{0, 0}, position{19, 5}, '╲'},
		{"just inside the diagonal, steep side", position{0, 0}, position{6, 5}, '╲'},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tunnelGlyph(tt.a, tt.b); got != tt.want {
				t.Errorf("tunnelGlyph(%v, %v) = %q, want %q", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// A tunnel includes both of its end cells, whichever way it is drawn.
func TestDrawTunnelIncludesBothEnds(t *testing.T) {
	cases := [][2]position{
		{{2, 3}, {7, 3}}, {{7, 3}, {2, 3}}, // horizontal, both ways
		{{4, 1}, {4, 8}}, {{4, 8}, {4, 1}}, // vertical
		{{0, 0}, {8, 4}}, {{8, 4}, {0, 0}}, // diagonal
		{{0, 6}, {8, 2}}, {{8, 2}, {0, 6}},
	}
	for _, tc := range cases {
		c := newCanvas(10, 10)
		c.drawTunnel(tc[0], tc[1])
		for _, p := range tc {
			if c.cells[p.Row][p.Col].st != styleTunnel {
				t.Errorf("%v to %v: the end %v is not drawn", tc[0], tc[1], p)
			}
		}
	}

	// A horizontal tunnel fills every cell from one end to the other.
	c := newCanvas(10, 5)
	c.drawTunnel(position{2, 3}, position{7, 3})
	for col := 2; col <= 7; col++ {
		if c.cells[3][col].r != '─' {
			t.Errorf("column %d of the tunnel is %q", col, c.cells[3][col].r)
		}
	}
}

// Two tunnels that cross make a '+' in the cell they share.
func TestDrawTunnelCrossing(t *testing.T) {
	c := newCanvas(10, 10)
	c.drawTunnel(position{0, 5}, position{9, 5})
	c.drawTunnel(position{4, 0}, position{4, 9})

	if got := c.cells[5][4].r; got != '+' {
		t.Errorf("crossing cell = %q, want '+'", got)
	}
	if got := c.cells[5][3].r; got != '─' {
		t.Errorf("a cell next to the crossing = %q, want '─'", got)
	}
	if got := c.cells[4][4].r; got != '│' {
		t.Errorf("a cell above the crossing = %q, want '│'", got)
	}
}

// Two tunnels with the same direction that share a cell do not make a '+'.
func TestDrawTunnelSameDirectionOverlap(t *testing.T) {
	c := newCanvas(10, 3)
	c.drawTunnel(position{0, 1}, position{6, 1})
	c.drawTunnel(position{3, 1}, position{9, 1})
	if got := c.cells[1][4].r; got != '─' {
		t.Errorf("shared cell = %q, want '─'", got)
	}
}

// Drawing partly or completely outside the grid does nothing and does not panic.
func TestDrawingOutsideTheGrid(t *testing.T) {
	c := newCanvas(5, 3)
	c.drawTunnel(position{-4, 1}, position{9, 1})
	c.drawTunnel(position{2, -3}, position{2, 8})
	c.drawLabel(position{3, 0}, "[long]", styleNone)
	c.drawLabel(position{0, 7}, "[x]", styleNone)
	c.drawLabel(position{-2, 1}, "[abcd]", styleNone)

	if got := c.text(false); !strings.Contains(got, "[l") {
		t.Errorf("the part of the label inside the grid is missing:\n%s", got)
	}
}

// With colours on, the start room is green, the end room red and a room with
// ants bold yellow. With colours off there is no escape code at all.
func TestRenderFrameColours(t *testing.T) {
	colony := graph.Colony{
		Ants: 2, Start: "s", End: "e",
		Rooms: []graph.Room{{Name: "s", X: 0, Y: 0}, {Name: "m", X: 5, Y: 0}, {Name: "e", X: 9, Y: 0}},
		Links: [][2]string{{"s", "m"}, {"m", "e"}},
	}
	occ := map[string][]int{"s": {2}, "m": {1}}
	l, err := newLayout(colony.Rooms, 30, 3)
	if err != nil {
		t.Fatal(err)
	}

	on := renderFrame(colony, l, occ, true)
	for _, want := range []string{
		ansiStart + "[s]" + ansiReset,
		ansiAnts + "[m]" + ansiReset,
		ansiEnd + "[e]" + ansiReset,
	} {
		if !strings.Contains(on, want) {
			t.Errorf("coloured frame has no %q:\n%q", want, on)
		}
	}

	if off := renderFrame(colony, l, occ, false); strings.Contains(off, "\x1b") {
		t.Errorf("a frame without colours contains an escape code: %q", off)
	}

	// A room with no ants is not coloured.
	empty := renderFrame(colony, l, map[string][]int{"s": {1, 2}}, true)
	if strings.Contains(empty, ansiAnts) {
		t.Error("a room without ants is coloured as if it had some")
	}
}

func TestAntsInRoomsLine(t *testing.T) {
	colony := graph.Colony{
		Start: "s", End: "e",
		Rooms: []graph.Room{{Name: "s"}, {Name: "m"}, {Name: "n"}, {Name: "e"}},
	}
	occ := map[string][]int{"s": {4}, "m": {1, 3}, "n": {2}, "e": {5}}

	// Room order of the file, not the order of the ants; start and end are not
	// listed because the status line counts them.
	if got, want := antsInRooms(colony, occ), "ants in rooms: [m] L1 L3  [n] L2"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := antsInRooms(colony, map[string][]int{"s": {1}}), "ants in rooms: none"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPlural(t *testing.T) {
	for n, want := range map[int]string{0: "0 ants", 1: "1 ant", 2: "2 ants", 1000: "1000 ants"} {
		if got := plural(n); got != want {
			t.Errorf("plural(%d) = %q, want %q", n, got, want)
		}
	}
}

// Every colony of testdata/ can be drawn: all room labels appear, no line has
// a space at the end, and the frame has the expected number of lines.
func TestRenderFrameOnTestdata(t *testing.T) {
	files, _ := filepath.Glob(filepath.Join("..", "..", "testdata", "*.txt"))
	checked := 0
	for _, file := range files {
		colony, _, err := parser.ParseFile(file)
		if err != nil {
			continue
		}
		l, err := newLayout(colony.Rooms, 78, 20)
		if err != nil {
			t.Errorf("%s: %v", file, err)
			continue
		}
		occ := newState(colony.Ants, colony.Start).occupancy()
		frame := renderFrame(colony, l, occ, false)

		for _, room := range colony.Rooms {
			if !strings.Contains(frame, label(room.Name)) {
				t.Errorf("%s: room %s is not in the frame", file, room.Name)
			}
		}
		lines := strings.Split(strings.TrimSuffix(frame, "\n"), "\n")
		if len(lines) != 20+2 {
			t.Errorf("%s: %d lines, want %d", file, len(lines), 22)
		}
		for _, line := range lines {
			if strings.HasSuffix(line, " ") {
				t.Errorf("%s: a line ends with a space: %q", file, line)
			}
		}
		checked++
	}
	if checked < 15 {
		t.Errorf("only %d files were checked", checked)
	}
}

// A room or a tunnel that the layout does not know is skipped, not a panic.
func TestRenderFrameSkipsWhatTheLayoutDoesNotKnow(t *testing.T) {
	colony := twoRooms()
	colony.Rooms = append(colony.Rooms, graph.Room{Name: "ghost", X: 5, Y: 5})
	colony.Links = append(colony.Links, [2]string{"s", "ghost"}, [2]string{"nobody", "e"})

	// The layout is built from the two real rooms only.
	l, err := newLayout(twoRooms().Rooms, 20, 3)
	if err != nil {
		t.Fatal(err)
	}
	frame := renderFrame(colony, l, nil, false)
	if strings.Contains(frame, "ghost") {
		t.Error("a room without a place was drawn")
	}
	if !strings.Contains(frame, "[s]") || !strings.Contains(frame, "[e]") {
		t.Error("the known rooms are missing")
	}
}

// A tunnel is a connected line with no gaps and no repeated cells: from the
// first cell to the last, each step moves at most one column and one row, and
// the line has exactly max(columns, rows) + 1 cells. Checked for every pair of
// cells of a small grid, in both directions.
func TestDrawTunnelIsAConnectedLine(t *testing.T) {
	const width, height = 12, 8
	for r1 := 0; r1 < height; r1++ {
		for c1 := 0; c1 < width; c1++ {
			for r2 := 0; r2 < height; r2++ {
				for c2 := 0; c2 < width; c2++ {
					a, b := position{c1, r1}, position{c2, r2}
					c := newCanvas(width, height)
					c.drawTunnel(a, b)

					count := 0
					for _, row := range c.cells {
						for _, cl := range row {
							if cl.st == styleTunnel {
								count++
							}
						}
					}
					if want := max(abs(c2-c1), abs(r2-r1)) + 1; count != want {
						t.Fatalf("%v to %v: %d cells, want %d", a, b, count, want)
					}
				}
			}
		}
	}
}
