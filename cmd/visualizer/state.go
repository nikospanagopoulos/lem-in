package main

import (
	"errors"
	"fmt"

	"lem-in/internal/sim"
)

// errUnknownAnt is returned when a move names an ant that the colony does not
// have (for example L7-a in a colony of 3 ants).
var errUnknownAnt = errors.New("unknown ant")

// state remembers where every ant is. At the start all of them are in the
// ##start room. Each turn moves some of them. The visualizer draws one frame
// per turn, so it calls apply once per turn and then asks who is in which room.
//
// Only the position of each ant is kept, and not the whole history. A move
// L2-5 says "ant 2 is now in room 5": it replaces the old position, so the
// state after the turn is enough to draw the frame.
type state struct {
	// room[i] is the name of the room where ant i+1 is. Ants are numbered from
	// 1 in the output, but a slice starts at 0, so ant n is stored at n-1.
	room []string
}

// newState puts all the ants in the start room.
func newState(ants int, start string) *state {
	room := make([]string, ants)
	for i := range room {
		room[i] = start
	}
	return &state{room: room}
}

// apply plays one turn: every ant in the turn goes to the room of its move. It
// does not check that the move is legal (sim.Check does that); it only needs
// the ant to exist. On an error the state is left as it was.
func (s *state) apply(turn sim.Turn) error {
	// First check every move, then change anything. That way a bad move in the
	// middle of a turn cannot leave the state half updated.
	for _, move := range turn {
		if move.Ant < 1 || move.Ant > len(s.room) {
			return fmt.Errorf("%w: ant %d, the colony has %d", errUnknownAnt, move.Ant, len(s.room))
		}
	}
	for _, move := range turn {
		s.room[move.Ant-1] = move.Room
	}
	return nil
}

// occupancy returns, for every room that has at least one ant, the numbers of
// its ants in increasing order. A room that is not a key of the map is empty:
// reading a missing key of a map gives the zero value, a nil slice, whose
// length is 0, so the caller does not need to check that the key exists.
func (s *state) occupancy() map[string][]int {
	rooms := make(map[string][]int)
	// The loop goes over the ants in order 1, 2, 3, ..., so each list is
	// already sorted and no sorting is needed.
	for i, room := range s.room {
		rooms[room] = append(rooms[room], i+1)
	}
	return rooms
}
