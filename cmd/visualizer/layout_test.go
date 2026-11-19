package main

import (
	"errors"
	"path/filepath"
	"testing"

	"lem-in/internal/graph"
	"lem-in/internal/parser"
)

func rooms(list ...graph.Room) []graph.Room { return list }

// Four rooms in the corners fill the whole area.
func TestLayoutCorners(t *testing.T) {
	l, err := newLayout(rooms(
		graph.Room{Name: "a", X: 0, Y: 0},
		graph.Room{Name: "b", X: 10, Y: 0},
		graph.Room{Name: "c", X: 0, Y: 5},
		graph.Room{Name: "d", X: 10, Y: 5},
	), 40, 10)
	if err != nil {
		t.Fatal(err)
	}

	// "[a]" is 3 columns wide, so a label can start at column 37 at most.
	want := map[string]position{
		"a": {0, 0}, "b": {37, 0},
		"c": {0, 9}, "d": {37, 9},
	}
	for name, w := range want {
		if got, _ := l.at(name); got != w {
			t.Errorf("room %s at %+v, want %+v", name, got, w)
		}
	}
}

// The order of the rooms is kept: a room with a smaller x is further left, and
// a room with a smaller y is higher.
func TestLayoutKeepsOrder(t *testing.T) {
	l, err := newLayout(rooms(
		graph.Room{Name: "a", X: 1, Y: 7},
		graph.Room{Name: "b", X: 50, Y: 3},
		graph.Room{Name: "c", X: 400, Y: 90},
	), 60, 20)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := l.at("a")
	b, _ := l.at("b")
	c, _ := l.at("c")
	if !(a.Col < b.Col && b.Col < c.Col) {
		t.Errorf("columns not in order: %d %d %d", a.Col, b.Col, c.Col)
	}
	if !(b.Row < a.Row && a.Row < c.Row) {
		t.Errorf("rows not in order: b=%d a=%d c=%d", b.Row, a.Row, c.Row)
	}
}

// Negative coordinates work: only the distances between rooms matter.
func TestLayoutNegativeCoordinates(t *testing.T) {
	l, err := newLayout(rooms(
		graph.Room{Name: "a", X: -100, Y: -50},
		graph.Room{Name: "b", X: 100, Y: 50},
	), 30, 10)
	if err != nil {
		t.Fatal(err)
	}
	if a, _ := l.at("a"); a != (position{0, 0}) {
		t.Errorf("a at %+v", a)
	}
	if b, _ := l.at("b"); b != (position{27, 9}) {
		t.Errorf("b at %+v", b)
	}
}

// Rooms on one line (all y equal, or all x equal) go to the middle of the
// other axis, with no division by zero.
func TestLayoutEqualCoordinates(t *testing.T) {
	l, err := newLayout(rooms(
		graph.Room{Name: "a", X: 0, Y: 4},
		graph.Room{Name: "b", X: 9, Y: 4},
	), 20, 11)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := l.at("a")
	b, _ := l.at("b")
	if a.Row != 5 || b.Row != 5 {
		t.Errorf("rows %d and %d, want both 5 (the middle)", a.Row, b.Row)
	}

	one, err := newLayout(rooms(graph.Room{Name: "x", X: 3, Y: 3}), 20, 11)
	if err != nil {
		t.Fatal(err)
	}
	if p, _ := one.at("x"); p != (position{8, 5}) {
		t.Errorf("a single room at %+v, want {8 5}", p)
	}
}

// Two rooms with the same coordinates must not be drawn over each other.
func TestLayoutSeparatesRoomsOnTheSameSpot(t *testing.T) {
	l, err := newLayout(rooms(
		graph.Room{Name: "a", X: 5, Y: 5},
		graph.Room{Name: "b", X: 5, Y: 5},
		graph.Room{Name: "c", X: 5, Y: 5},
		graph.Room{Name: "far", X: 50, Y: 50},
	), 40, 10)
	if err != nil {
		t.Fatal(err)
	}
	assertNoOverlap(t, l)

	// The first room keeps its place, and the same input always gives the
	// same layout.
	again, _ := newLayout(rooms(
		graph.Room{Name: "a", X: 5, Y: 5},
		graph.Room{Name: "b", X: 5, Y: 5},
		graph.Room{Name: "c", X: 5, Y: 5},
		graph.Room{Name: "far", X: 50, Y: 50},
	), 40, 10)
	for name, p := range l.rooms {
		if again.rooms[name] != p {
			t.Errorf("room %s moved between two runs", name)
		}
	}
}

// With a long label the room must not go outside the area.
func TestLayoutLongNameStaysInside(t *testing.T) {
	l, err := newLayout(rooms(
		graph.Room{Name: "a", X: 0, Y: 0},
		graph.Room{Name: "gilfoyle", X: 100, Y: 100},
	), 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	p, _ := l.at("gilfoyle")
	if p.Col+len("[gilfoyle]") > 30 {
		t.Errorf("label ends at column %d, outside the 30 columns", p.Col+len("[gilfoyle]"))
	}
}

func TestLayoutCentre(t *testing.T) {
	l, _ := newLayout(rooms(graph.Room{Name: "abcd", X: 0, Y: 0}, graph.Room{Name: "z", X: 9, Y: 9}), 40, 10)
	// "[abcd]" is 6 columns wide, so its centre is 3 columns from the "[".
	c, ok := l.centre("abcd")
	if !ok || c != (position{3, 0}) {
		t.Errorf("centre = %+v, %v", c, ok)
	}
	if _, ok := l.centre("nobody"); ok {
		t.Error("an unknown room has a centre")
	}
	if _, ok := l.at("nobody"); ok {
		t.Error("an unknown room has a position")
	}
}

func TestLayoutTooSmall(t *testing.T) {
	two := rooms(graph.Room{Name: "longname", X: 0, Y: 0}, graph.Room{Name: "b", X: 1, Y: 1})
	for _, size := range []struct{ w, h int }{{5, 10}, {40, 0}, {40, -3}} {
		if _, err := newLayout(two, size.w, size.h); !errors.Is(err, errTooSmall) {
			t.Errorf("%dx%d: error = %v, want errTooSmall", size.w, size.h, err)
		}
	}
}

func TestLayoutNoRooms(t *testing.T) {
	if _, err := newLayout(nil, 10, 10); err != nil {
		t.Fatalf("no rooms: %v", err)
	}
}

// Every colony of testdata/, on a normal terminal: every room gets a place
// inside the area and no two labels overlap.
func TestLayoutOnTestdata(t *testing.T) {
	files, _ := filepath.Glob(filepath.Join("..", "..", "testdata", "*.txt"))
	checked := 0
	for _, file := range files {
		colony, _, err := parser.ParseFile(file)
		if err != nil {
			continue // the bad files have no colony
		}
		l, err := newLayout(colony.Rooms, 78, 22)
		if err != nil {
			t.Errorf("%s: %v", file, err)
			continue
		}
		for _, room := range colony.Rooms {
			p, ok := l.at(room.Name)
			w := l.labelWidth[room.Name]
			if !ok || p.Col < 0 || p.Col+w > 78 || p.Row < 0 || p.Row >= 22 {
				t.Errorf("%s: room %s at %+v is outside the area", file, room.Name, p)
			}
		}
		assertNoOverlap(t, l)
		checked++
	}
	if checked < 15 {
		t.Errorf("only %d files were checked", checked)
	}
}

// assertNoOverlap fails if two labels touch on the same row.
func assertNoOverlap(t *testing.T, l layout) {
	t.Helper()
	for a, pa := range l.rooms {
		for b, pb := range l.rooms {
			if a >= b || pa.Row != pb.Row {
				continue
			}
			sa := span{pa.Col, pa.Col + l.labelWidth[a]}
			sb := span{pb.Col, pb.Col + l.labelWidth[b]}
			if sa.overlaps(sb) {
				t.Errorf("rooms %s and %s overlap on row %d", a, b, pa.Row)
			}
		}
	}
}

// When the area is so full that a room fits nowhere, the layout must still
// finish and give every room a place inside the area (labels may then overlap).
func TestLayoutFullAreaDoesNotFail(t *testing.T) {
	l, err := newLayout(rooms(
		graph.Room{Name: "a", X: 1, Y: 1},
		graph.Room{Name: "b", X: 1, Y: 1},
		graph.Room{Name: "c", X: 1, Y: 1},
	), 5, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(l.rooms) != 3 {
		t.Fatalf("%d rooms placed, want 3", len(l.rooms))
	}
	for name, p := range l.rooms {
		if p.Row != 0 || p.Col < 0 || p.Col+3 > 5 {
			t.Errorf("room %s at %+v is outside the area", name, p)
		}
	}
}

// The rows are spread over the whole height, from row 0 to the last row.
func TestLayoutRowScale(t *testing.T) {
	l, err := newLayout(rooms(
		graph.Room{Name: "a", X: 0, Y: 0},
		graph.Room{Name: "b", X: 5, Y: 7},
		graph.Room{Name: "c", X: 9, Y: 10},
	), 40, 10)
	if err != nil {
		t.Fatal(err)
	}
	// y=7 of 0..10 is 70% of the way: 0.7 * 9 = 6.3, which rounds to row 6.
	for name, want := range map[string]int{"a": 0, "b": 6, "c": 9} {
		if p, _ := l.at(name); p.Row != want {
			t.Errorf("room %s in row %d, want %d", name, p.Row, want)
		}
	}
}

// Two labels that would touch are not drawn side by side: one column stays
// free, or the second room moves to another row.
func TestLayoutKeepsAGapBetweenLabels(t *testing.T) {
	// Both rooms want row 1 of 3. With width 6, "[a]" and "[b]" would land in
	// columns 0-2 and 3-5, touching each other.
	l, err := newLayout(rooms(
		graph.Room{Name: "a", X: 0, Y: 0},
		graph.Room{Name: "b", X: 1, Y: 0},
	), 6, 3)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := l.at("a")
	b, _ := l.at("b")
	if a.Row == b.Row {
		t.Errorf("both rooms are in row %d, touching", a.Row)
	}
}

// The area must be at least as wide as the longest label, and exactly that
// width is enough.
func TestLayoutWidthLimit(t *testing.T) {
	two := rooms(graph.Room{Name: "ab", X: 0, Y: 0}, graph.Room{Name: "c", X: 1, Y: 1})
	// "[ab]" needs 4 columns.
	if _, err := newLayout(two, 3, 5); !errors.Is(err, errTooSmall) {
		t.Errorf("width 3: error = %v, want errTooSmall", err)
	}
	if _, err := newLayout(two, 4, 5); err != nil {
		t.Errorf("width 4: %v", err)
	}
}
