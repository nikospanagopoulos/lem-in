package sim

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"lem-in/internal/graph"
)

// colony builds a colony for a test. Tunnels are written as "a-b". Check never
// looks at the list of rooms, so we leave it empty to keep the tests short.
func colony(ants int, start, end string, tunnels ...string) graph.Colony {
	c := graph.Colony{Ants: ants, Start: start, End: end}
	for _, t := range tunnels {
		a, b, _ := strings.Cut(t, "-") // cut at the first dash
		c.Links = append(c.Links, [2]string{a, b})
	}
	return c
}

// turns builds a list of turns from text written exactly like the program's
// output: one string per turn, moves separated by spaces, for example
// "L1-a L2-b". Reading the tests is then like reading a real output.
func turns(lines ...string) []Turn {
	var all []Turn
	for _, line := range lines {
		var turn Turn
		for _, word := range strings.Fields(line) {
			ant, room, _ := strings.Cut(strings.TrimPrefix(word, "L"), "-")
			n, err := strconv.Atoi(ant)
			if err != nil {
				panic("bad test input: " + word)
			}
			turn = append(turn, Move{Ant: n, Room: room})
		}
		all = append(all, turn)
	}
	return all
}

// The colony of Example 1 of the subject (3 ants, start "1", end "0").
func subjectColony() graph.Colony {
	return colony(3, "1", "0",
		"0-4", "0-6", "1-3", "4-3", "5-2", "3-5",
		"4-2", "2-1", "7-6", "7-2", "7-4", "6-5")
}

// Runs that follow every rule must give no error.
func TestCheckAcceptsLegalRuns(t *testing.T) {
	tests := []struct {
		name  string
		c     graph.Colony
		turns []Turn
	}{
		{
			// The output printed in the subject.
			name: "subject example 1",
			c:    subjectColony(),
			turns: turns(
				"L1-3 L2-2",
				"L1-4 L2-5 L3-3",
				"L1-0 L2-6 L3-4",
				"L2-0 L3-0",
			),
		},
		{
			// Same idea as audit example02, turn 2: an ant leaves a room
			// and another ant enters that room in the SAME turn. This is
			// legal, because rooms are judged at the end of the turn.
			name: "a room emptied in a turn can be entered in the same turn",
			c:    colony(2, "s", "e", "s-a", "a-b", "b-e"),
			turns: turns(
				"L1-a",
				"L1-b L2-a",
				"L1-e L2-b",
				"L2-e",
			),
		},
		{
			// Start and end may hold any number of ants: two ants reach
			// the end room in the same turn, from two different rooms.
			name: "end room can receive several ants in one turn",
			c:    colony(2, "s", "e", "s-a", "s-b", "a-e", "b-e"),
			turns: turns(
				"L1-a L2-b",
				"L1-e L2-e",
			),
		},
		{
			// The start room may hold any number of ants, also when an ant
			// walks back into it: after turn 2 it holds both ants again.
			// (A silly run, but legal under the rules.)
			name: "start room can receive an ant that walks back",
			c:    colony(2, "s", "e", "s-a", "a-e"),
			turns: turns(
				"L1-a",
				"L1-s",
				"L1-a",
				"L1-e L2-a",
				"L2-e",
			),
		},
		{
			// Tunnels work both ways: "b-a" in the file allows a to b.
			name:  "a tunnel can be used in both directions",
			c:     colony(1, "s", "e", "s-a", "e-a"),
			turns: turns("L1-a", "L1-e"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := Check(tc.c, tc.turns); err != nil {
				t.Fatalf("Check returned an error: %v", err)
			}
		})
	}
}

// One bad run for every rule. errors.Is checks WHICH rule was found broken, not
// the wording of the message. The "want" column is the rule that must be named.
func TestCheckRejectsEveryRule(t *testing.T) {
	chain := colony(2, "s", "e", "s-a", "a-b", "b-e") // s - a - b - e
	diamond := colony(2, "s", "e", "s-a", "s-b", "a-c", "b-c", "c-e")

	tests := []struct {
		name  string
		c     graph.Colony
		turns []Turn
		want  error
	}{
		{
			name:  "ant number zero",
			c:     chain,
			turns: turns("L0-a"),
			want:  ErrUnknownAnt,
		},
		{
			name:  "ant number above the number of ants",
			c:     chain,
			turns: turns("L3-a"),
			want:  ErrUnknownAnt,
		},
		{
			name:  "an ant moves twice in one turn",
			c:     chain,
			turns: turns("L1-a L1-b"),
			want:  ErrAntMovedTwice,
		},
		{
			name:  "an ant leaves the end room",
			c:     colony(1, "s", "e", "s-e"),
			turns: turns("L1-e", "L1-s"),
			want:  ErrLeftEnd,
		},
		{
			name:  "a move between two rooms with no tunnel",
			c:     chain,
			turns: turns("L1-b"), // s and b are not linked
			want:  ErrNoTunnel,
		},
		{
			name:  "a move to a room that does not exist",
			c:     chain,
			turns: turns("L1-nowhere"),
			want:  ErrNoTunnel,
		},
		{
			// The ant is in "a" after turn 1, so "a-e" is not a move it
			// can make, even though the tunnel s-a exists.
			name:  "a move that does not start where the ant is",
			c:     chain,
			turns: turns("L1-a", "L1-e"),
			want:  ErrNoTunnel,
		},
		{
			name:  "two ants use the same tunnel in one turn",
			c:     chain,
			turns: turns("L1-a L2-a"),
			want:  ErrTunnelReused,
		},
		{
			// Ant 1 goes a to b while ant 2 goes b to a: they swap. Each
			// room still has one ant at the end of the turn, so only the
			// tunnel rule can catch it.
			name: "two ants swap places through one tunnel",
			c:    colony(2, "s", "e", "s-a", "a-b", "b-e"),
			turns: turns(
				"L1-a",
				"L1-b L2-a",
				"L1-a L2-b",
			),
			want: ErrTunnelReused,
		},
		{
			// Two different tunnels lead into room c in the same turn.
			name: "two ants end the turn in the same room",
			c:    diamond,
			turns: turns(
				"L1-a L2-b",
				"L1-c L2-c",
			),
			want: ErrRoomFull,
		},
		{
			// Ant 1 stays in "a", so ant 2 cannot enter it in turn 2.
			name: "an ant enters a room whose ant did not leave",
			c:    colony(2, "s", "e", "s-a", "a-e"),
			turns: turns(
				"L1-a",
				"L2-a",
			),
			want: ErrRoomFull,
		},
		{
			name:  "the run ends before every ant arrived",
			c:     chain,
			turns: turns("L1-a", "L1-b L2-a", "L1-e"),
			want:  ErrNotFinished,
		},
		{
			name:  "no turns at all",
			c:     chain,
			turns: nil,
			want:  ErrNotFinished,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Check(tc.c, tc.turns)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got error %v, want %v", err, tc.want)
			}
		})
	}
}

// The card asks for errors that name the turn and the ant, so a person reading
// the message knows where to look.
func TestCheckErrorNamesTurnAndAnt(t *testing.T) {
	c := colony(2, "s", "e", "s-a", "a-e")
	err := Check(c, turns("L1-a", "L2-a"))
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, part := range []string{"turn 2", "ant 2"} {
		if !strings.Contains(err.Error(), part) {
			t.Fatalf("message %q does not contain %q", err, part)
		}
	}
}

// A colony without ants cannot be replayed. This must be an error, not a panic.
func TestCheckRejectsColonyWithoutAnts(t *testing.T) {
	for _, ants := range []int{0, -1} {
		if err := Check(colony(ants, "s", "e", "s-e"), nil); err == nil {
			t.Fatalf("expected an error for %d ants", ants)
		}
	}
}

// Check must only read its arguments.
func TestCheckDoesNotChangeItsArguments(t *testing.T) {
	c := subjectColony()
	in := turns("L1-3 L2-2", "L1-4 L2-5 L3-3", "L1-0 L2-6 L3-4", "L2-0 L3-0")
	before := turns("L1-3 L2-2", "L1-4 L2-5 L3-3", "L1-0 L2-6 L3-4", "L2-0 L3-0")

	if err := Check(c, in); err != nil {
		t.Fatal(err)
	}
	for i := range in {
		for j := range in[i] {
			if in[i][j] != before[i][j] {
				t.Fatalf("turn %d, move %d changed: %v", i+1, j, in[i][j])
			}
		}
	}
}
