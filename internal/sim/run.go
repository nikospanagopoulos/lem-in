package sim

import "lem-in/internal/paths"

// walker is an ant that has left the start room and is now walking along its
// path. It remembers the two things we need to know to move it: which path it
// is on, and where it is on that path.
type walker struct {
	ant  int // ant number, from 1
	path int // index of its path in the set
	at   int // index in the path of the room the ant is in now (0 = start room)
}

// Run plays the whole movement and returns every turn, from the first turn
// until the last ant is in the end room. Each Turn lists the moves of that turn
// in order of ant number, so it can be printed as it is.
//
// Run uses Assign to know how many ants go down each path. Then, in every turn:
//
//  1. every ant that is already walking moves one room forward;
//  2. every path that still has ants waiting in the start room sends the next
//     one into its first room.
//
// Ants are numbered in the order they leave the start room. When two paths send
// an ant in the same turn, the shorter path (it comes first in the set) gets
// the smaller number.
//
// Why we never check for collisions here: the paths of a set share no room
// except start and end (that is what paths.Sets promises), and ants on the same
// path are always one room apart, because a path sends one ant per turn and
// every ant moves one room per turn. So two ants can never be in the same room,
// and a tunnel is never used twice in one turn. Check, written separately,
// proves it in the tests.
//
// set must be sorted shortest first, and every path must have at least two
// rooms (start and end are different rooms; the parser guarantees it). Run
// returns nil when there is nothing to move.
func Run(set []paths.Path, ants int) []Turn {
	if len(set) == 0 || ants <= 0 {
		return nil
	}

	// waiting[i] is how many ants still have to leave the start room on path i.
	waiting := Assign(set, ants)

	var turns []Turn
	var walking []walker // ants on a path, always in increasing ant number
	notLeft := ants      // ants still in the start room
	nextAnt := 1         // number of the next ant to leave the start room

	// The run ends when nobody is waiting and nobody is walking. We do not
	// count the turns in advance: the loop itself finds when it is over.
	for notLeft > 0 || len(walking) > 0 {
		var turn Turn

		// Step 1: move the ants that are already on a path. "still" reuses
		// the memory of "walking" (it only ever writes at a place we have
		// already read), and keeps the ants that have not arrived yet.
		still := walking[:0]
		for _, w := range walking {
			w.at++ // w is a copy, so we change it and keep the copy below
			turn = append(turn, Move{Ant: w.ant, Room: set[w.path][w.at]})
			if w.at < len(set[w.path])-1 { // not in the end room yet
				still = append(still, w)
			}
		}
		walking = still

		// Step 2: each path that has ants waiting sends one into its first
		// room. We go through the paths shortest first.
		for i, path := range set {
			if waiting[i] == 0 {
				continue
			}
			waiting[i]--
			notLeft--

			turn = append(turn, Move{Ant: nextAnt, Room: path[1]})
			// A path of one tunnel leads straight into the end room: the
			// ant has already arrived and does not walk any further.
			if len(path) > 2 {
				walking = append(walking, walker{ant: nextAnt, path: i, at: 1})
			}
			nextAnt++
		}

		// The moves are sorted by ant number without sorting them: the ants
		// that were already walking have smaller numbers than the ants that
		// just left, and "walking" keeps its ants in order. The tests check
		// it.
		turns = append(turns, turn)
	}
	return turns
}
