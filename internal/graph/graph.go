// Package graph holds the types shared by every other package: the colony as
// read from the input file, and the graph of rooms and tunnels built from it.
//
// Two types, two jobs:
//   - Colony is plain data: what the input file said. The parser produces it.
//   - Graph is built from a Colony and answers one question: "which rooms are
//     next to this room?". The path-finding code asks that question again and
//     again, so the Graph stores the answer ready to use.
package graph

import "slices"

// Room is one room of the colony. X and Y are only used by the visualizer;
// the solver never looks at them.
type Room struct {
	Name string
	X, Y int
}

// Colony is the parsed content of an input file. The parser produces it and
// every other package reads it. Rooms and Links keep the order of the file, so
// the program gives the same result on every run.
type Colony struct {
	Ants  int         // number of ants, always greater than 0
	Start string      // name of the ##start room
	End   string      // name of the ##end room
	Rooms []Room      // all rooms, including start and end
	Links [][2]string // tunnels, each one the names of its two rooms
}

// Graph answers "which rooms are next to this room?". Build it with New.
//
// Its fields start with a lowercase letter, so only this package can touch
// them. Other packages must use the methods, which means nobody can put the
// graph in an inconsistent state by accident.
type Graph struct {
	rooms []string            // room names, in file order
	adj   map[string][]string // room name -> names of the rooms it is connected to
}

// New builds the graph of a colony. Tunnels are two-way. New does not change
// the colony it receives.
func New(c Colony) *Graph {
	g := &Graph{
		// make with a capacity reserves the space once, instead of growing
		// the slice and the map step by step.
		rooms: make([]string, 0, len(c.Rooms)),
		adj:   make(map[string][]string, len(c.Rooms)),
	}

	for _, room := range c.Rooms {
		g.rooms = append(g.rooms, room.Name)
	}

	// A tunnel works in both directions, so we record it twice: b is a
	// neighbour of a, and a is a neighbour of b. Reading a missing key of a
	// map gives a nil slice, and append works on a nil slice, so we do not
	// need to create each list first.
	for _, link := range c.Links {
		a, b := link[0], link[1]
		g.adj[a] = append(g.adj[a], b)
		g.adj[b] = append(g.adj[b], a)
	}
	return g
}

// Rooms returns the names of all rooms, in file order. The caller may modify
// the returned slice: it is a copy.
func (g *Graph) Rooms() []string {
	// If we returned g.rooms itself, the caller would hold the graph's own
	// slice and could change the graph by writing into it. A copy is safe.
	return slices.Clone(g.rooms)
}

// Neighbours returns the names of the rooms connected to room, in the order of
// the tunnels in the file. It returns an empty result for a room that has no
// tunnels and for a room that does not exist. The caller may modify the
// returned slice: it is a copy.
func (g *Graph) Neighbours(room string) []string {
	// A room that is not in the map gives a nil slice, and cloning nil gives
	// nil. So "unknown room" and "room with no tunnels" both return an empty
	// result, with no special case and no panic.
	return slices.Clone(g.adj[room])
}
