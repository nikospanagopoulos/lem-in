package graph

import (
	"reflect"
	"slices"
	"testing"
)

// sample is a small colony used by all the tests below:
//
//	s - a - b - e      plus a shortcut s - b, and a room "iso" with no tunnel
func sample() Colony {
	return Colony{
		Ants:  1,
		Start: "s",
		End:   "e",
		Rooms: []Room{
			{Name: "s"}, {Name: "a", X: 1}, {Name: "b", X: 2},
			{Name: "e", X: 3}, {Name: "iso", X: 9, Y: 9},
		},
		Links: [][2]string{{"s", "a"}, {"a", "b"}, {"b", "e"}, {"s", "b"}},
	}
}

// Rooms must come back in the order of the file, not in any other order.
func TestRoomsKeepsFileOrder(t *testing.T) {
	got := New(sample()).Rooms()
	want := []string{"s", "a", "b", "e", "iso"}
	if !slices.Equal(got, want) {
		t.Fatalf("Rooms() = %v, want %v", got, want)
	}
}

// A "table-driven" test: the cases are rows of a table, and one loop runs them
// all. To test a new case, add a row. t.Run gives each row its own name, so a
// failure tells you which row broke.
func TestNeighbours(t *testing.T) {
	g := New(sample())
	tests := []struct {
		name string
		room string
		want []string
	}{
		{"start room", "s", []string{"a", "b"}},
		{"middle room", "a", []string{"s", "b"}},
		{"end room", "e", []string{"b"}},
		{"isolated room", "iso", nil},
		{"unknown room", "nope", nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := g.Neighbours(tc.room)
			// slices.Equal treats nil and an empty slice as equal, which is
			// what we want: "no neighbours" is one answer, not two.
			if !slices.Equal(got, tc.want) {
				t.Fatalf("Neighbours(%q) = %v, want %v", tc.room, got, tc.want)
			}
		})
	}
}

// The neighbours of a room must come in the same order on every call.
// This matters because Go's map order is random: if New or Neighbours
// depended on it, the program could print different moves on different runs.
func TestNeighboursOrderIsStable(t *testing.T) {
	g := New(sample())
	first := g.Neighbours("b")
	for i := 0; i < 50; i++ {
		if got := g.Neighbours("b"); !slices.Equal(got, first) {
			t.Fatalf("run %d: Neighbours(b) = %v, first call gave %v", i, got, first)
		}
	}
}

// If a caller writes into a returned slice, the graph must not change.
func TestReturnedSlicesAreCopies(t *testing.T) {
	g := New(sample())

	g.Neighbours("s")[0] = "changed"
	if got := g.Neighbours("s"); got[0] != "a" {
		t.Fatalf("changing a result of Neighbours changed the graph: %v", got)
	}

	g.Rooms()[0] = "changed"
	if got := g.Rooms(); got[0] != "s" {
		t.Fatalf("changing a result of Rooms changed the graph: %v", got)
	}
}

// Building a graph must only read the colony, never modify it.
func TestNewDoesNotChangeTheColony(t *testing.T) {
	c := sample()
	before := sample() // a second, separate copy to compare against
	New(c)
	// reflect.DeepEqual compares slices element by element, not by address.
	if !reflect.DeepEqual(c, before) {
		t.Fatalf("New changed its argument: %+v, want %+v", c, before)
	}
}
