package sim

import (
	"errors"
	"fmt"

	"lem-in/internal/graph"
)

// The rules Check can find broken. Each one is a value created once with
// errors.New, so a caller (or a test) recognises the rule with errors.Is and
// does not depend on the wording of the message.
//
// The error that Check returns also carries the turn number, the ant and the
// rooms involved. It wraps one of these values with %w, so errors.Is works
// on it.
var (
	ErrUnknownAnt    = errors.New("ant number out of range")
	ErrAntMovedTwice = errors.New("ant moved more than once in a turn")
	ErrLeftEnd       = errors.New("ant left the end room")
	ErrNoTunnel      = errors.New("move does not follow a tunnel")
	ErrTunnelReused  = errors.New("tunnel used more than once in a turn")
	ErrRoomFull      = errors.New("room holds more than one ant")
	ErrNotFinished   = errors.New("not all ants reached the end room")
)

// Check replays a list of turns on a colony and returns the first rule that is
// broken, or nil if the whole run is legal. The rules are those of section 5 of
// the PRD:
//
//   - an ant moves at most once per turn, and never leaves the end room;
//   - every move follows a tunnel, starting from the room the ant is in;
//   - a tunnel is used at most once per turn;
//   - at the END of a turn, no room except start and end holds two ants;
//   - at the end of the run, every ant is in the end room.
//
// Check is written to be INDEPENDENT of the code that produces the turns. It
// does not know the paths and does not use Run or Assign. It sees only what an
// auditor sees: the colony and the moves. A checker that shared code with the
// solver could share its bugs, and then it would agree with a wrong answer.
//
// Check only reads c.Ants, c.Start, c.End and c.Links, and it does not change
// its arguments.
func Check(c graph.Colony, turns []Turn) error {
	if c.Ants < 1 {
		// Also protects the make call below, which panics on a negative size.
		return fmt.Errorf("check: colony has %d ants", c.Ants)
	}

	// Look-up tables, built once, so each move costs the same however big the
	// colony is.
	tunnels := tunnelSet(c.Links)

	// position[ant] is the room the ant is in right now. Index 0 is unused:
	// ants are numbered from 1, so we make the slice one longer instead of
	// writing "ant-1" everywhere.
	position := make([]string, c.Ants+1)
	for ant := 1; ant <= c.Ants; ant++ {
		position[ant] = c.Start // all ants begin in the start room
	}

	// occupants[room] counts the ants in a room right now. Reading a room that
	// is not in the map gives 0, which is exactly "nobody there".
	occupants := map[string]int{c.Start: c.Ants}

	for i, turn := range turns {
		number := i + 1 // turns are numbered from 1 in messages

		// What this turn has already used. They start empty on every turn,
		// because "once per turn" means the memory is erased between turns.
		moved := make(map[int]bool, len(turn))
		used := make(map[[2]string]bool, len(turn))

		for _, move := range turn {
			ant := move.Ant
			if ant < 1 || ant > c.Ants {
				return violation(number, ant, ErrUnknownAnt, fmt.Sprintf("there are %d ants", c.Ants))
			}
			if moved[ant] {
				return violation(number, ant, ErrAntMovedTwice, "")
			}
			moved[ant] = true

			from := position[ant]
			if from == c.End {
				// Following a tunnel out of the end room would be legal for
				// the tunnel rule below, so this rule needs its own check.
				return violation(number, ant, ErrLeftEnd, fmt.Sprintf("tried to go to %q", move.Room))
			}

			key := tunnelKey(from, move.Room)
			if !tunnels[key] {
				// This also catches a room that does not exist, because no
				// tunnel in the colony mentions it.
				return violation(number, ant, ErrNoTunnel, fmt.Sprintf("%q to %q", from, move.Room))
			}
			if used[key] {
				return violation(number, ant, ErrTunnelReused, fmt.Sprintf("%q to %q", from, move.Room))
			}
			used[key] = true

			// The move is legal so far: update where the ant is. We only
			// COUNT here. Whether a room is too full is judged after all the
			// moves of the turn (below), because a room that one ant leaves
			// can be entered by another ant in the same turn.
			position[ant] = move.Room
			occupants[from]--
			occupants[move.Room]++
		}

		// End of the turn: now look at the rooms that received an ant.
		// A room can only become too full by receiving someone, so these are
		// the only rooms worth checking. Start and end may hold any number.
		for _, move := range turn {
			room := move.Room
			if room == c.Start || room == c.End {
				continue
			}
			if occupants[room] > 1 {
				return violation(number, move.Ant, ErrRoomFull,
					fmt.Sprintf("room %q has %d ants", room, occupants[room]))
			}
		}
	}

	if inEnd := occupants[c.End]; inEnd != c.Ants {
		return fmt.Errorf("%w: %d of %d ants are in the end room", ErrNotFinished, inEnd, c.Ants)
	}
	return nil
}

// violation builds the error that Check returns. It says which turn and which
// ant, and it wraps the rule so that errors.Is can find it.
func violation(turn, ant int, rule error, detail string) error {
	if detail == "" {
		return fmt.Errorf("turn %d, ant %d: %w", turn, ant, rule)
	}
	return fmt.Errorf("turn %d, ant %d: %w (%s)", turn, ant, rule, detail)
}

// tunnelKey identifies a tunnel by its two rooms, whichever way it is written:
// "a-b" and "b-a" give the same key. An array like [2]string can be a map key
// (a slice cannot), which is why the key is an array.
func tunnelKey(a, b string) [2]string {
	if a > b {
		a, b = b, a
	}
	return [2]string{a, b}
}

// tunnelSet turns the list of tunnels into a set for fast look-up.
func tunnelSet(links [][2]string) map[[2]string]bool {
	set := make(map[[2]string]bool, len(links))
	for _, link := range links {
		set[tunnelKey(link[0], link[1])] = true
	}
	return set
}
