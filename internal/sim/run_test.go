package sim

import (
	"fmt"
	"math/rand"
	"reflect"
	"slices"
	"testing"

	"lem-in/internal/graph"
	"lem-in/internal/paths"
)

// named builds a path from room names, written start first.
func named(rooms ...string) paths.Path { return paths.Path(rooms) }

// withRooms fills in the list of rooms of a colony from its tunnels, in the
// order they first appear. The colony helper of check_test.go leaves it empty,
// because Check does not need it, but paths.Sets does.
func withRooms(c graph.Colony) graph.Colony {
	seen := map[string]bool{}
	add := func(name string) {
		if !seen[name] {
			seen[name] = true
			c.Rooms = append(c.Rooms, graph.Room{Name: name})
		}
	}
	add(c.Start)
	add(c.End)
	for _, link := range c.Links {
		add(link[0])
		add(link[1])
	}
	return c
}

// The hand exercise of the card. Path A has 3 tunnels, path B has 4, there are
// 3 ants. Assign gives [2, 1]. The turns are the same shape as the output of the
// subject (Example 1): ant 2 takes the long path, ant 3 follows ant 1.
func TestRunHandExample(t *testing.T) {
	set := []paths.Path{
		named("s", "a", "b", "e"),
		named("s", "c", "d", "f", "e"),
	}
	want := turns(
		"L1-a L2-c",
		"L1-b L2-d L3-a",
		"L1-e L2-f L3-b",
		"L2-e L3-e",
	)
	got := Run(set, 3)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("turns:\n got %v\nwant %v", got, want)
	}
}

// A path of one tunnel leads straight from start to end: one ant per turn.
func TestRunDirectTunnel(t *testing.T) {
	got := Run([]paths.Path{named("s", "e")}, 3)
	want := turns("L1-e", "L2-e", "L3-e")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("turns:\n got %v\nwant %v", got, want)
	}
}

// Nothing to move: no turns, and no panic.
func TestRunNothingToMove(t *testing.T) {
	if got := Run(nil, 5); got != nil {
		t.Fatalf("no paths: got %v", got)
	}
	if got := Run([]paths.Path{named("s", "a", "e")}, 0); got != nil {
		t.Fatalf("no ants: got %v", got)
	}
}

// colonyOfSet builds a colony in which the paths of the set are the only
// tunnels. Rooms are shared only by start "s" and end "e", like the paths of a
// set from paths.Sets. Check can then replay the turns on this colony.
func colonyOfSet(set []paths.Path, ants int) graph.Colony {
	c := graph.Colony{Ants: ants, Start: "s", End: "e"}
	for _, p := range set {
		for i := 1; i < len(p); i++ {
			c.Links = append(c.Links, [2]string{p[i-1], p[i]})
		}
	}
	return withRooms(c)
}

// randomSet builds a set of disjoint paths, shortest first. At most one path
// has a single tunnel, because two tunnels from "s" to "e" would be the same
// tunnel written twice, and the parser rejects that.
func randomSet(rng *rand.Rand) []paths.Path {
	k := 1 + rng.Intn(4)
	lengths := make([]int, k)
	direct := false
	for i := range lengths {
		lengths[i] = 1 + rng.Intn(8)
		if lengths[i] == 1 {
			if direct {
				lengths[i] = 2
			}
			direct = true
		}
	}
	slices.Sort(lengths)

	set := make([]paths.Path, k)
	for i, l := range lengths {
		p := paths.Path{"s"}
		for j := 1; j < l; j++ {
			p = append(p, fmt.Sprintf("p%d_%d", i, j)) // a room used by this path only
		}
		set[i] = append(p, "e")
	}
	return set
}

// The big test. For many random sets and numbers of ants, the result of Run
// must satisfy everything the card asks for:
//   - it follows every rule (the independent Check says nothing);
//   - it has the number of turns that paths.Turns predicts;
//   - no turn is empty, and every turn is sorted by ant number;
//   - every ant moves, and ends in the end room.
func TestRunRandomSets(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	for i := 0; i < 5000; i++ {
		set := randomSet(rng)
		ants := 1 + rng.Intn(40)
		c := colonyOfSet(set, ants)

		got := Run(set, ants)

		if err := Check(c, got); err != nil {
			t.Fatalf("set %v, %d ants: Check says %v", set, ants, err)
		}
		if want := paths.Turns(set, ants); len(got) != want {
			t.Fatalf("set %v, %d ants: %d turns, paths.Turns says %d", set, ants, len(got), want)
		}
		for n, turn := range got {
			if len(turn) == 0 {
				t.Fatalf("set %v, %d ants: turn %d is empty", set, ants, n+1)
			}
			sorted := slices.IsSortedFunc(turn, func(a, b Move) int { return a.Ant - b.Ant })
			if !sorted {
				t.Fatalf("set %v, %d ants: turn %d is not sorted: %v", set, ants, n+1, turn)
			}
		}
	}
}

// The whole chain on real colonies: Sets finds the paths, Best chooses, Run
// moves the ants, and Check approves. The numbers of turns are the ones of the
// PRD tables.
func TestRunWithRealPaths(t *testing.T) {
	tests := []struct {
		name      string
		colony    graph.Colony
		wantTurns int
	}{
		{"subject example 1", withRooms(subjectColony()), 4},
		// Audit example02 (20 ants): a direct tunnel and one longer path.
		{"example02", withRooms(colony(20, "0", "3", "0-1", "0-3", "1-2", "3-2")), 11},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.colony
			sets, err := paths.Sets(graph.New(c), c.Start, c.End)
			if err != nil {
				t.Fatal(err)
			}
			best := paths.Best(sets, c.Ants)

			got := Run(best, c.Ants)

			if err := Check(c, got); err != nil {
				t.Fatalf("Check says %v", err)
			}
			if len(got) != tc.wantTurns {
				t.Fatalf("%d turns, want %d", len(got), tc.wantTurns)
			}
		})
	}
}

// Run must only read the set it receives.
func TestRunDoesNotChangeTheSet(t *testing.T) {
	set := []paths.Path{named("s", "a", "e"), named("s", "b", "c", "e")}
	before := []paths.Path{named("s", "a", "e"), named("s", "b", "c", "e")}

	Run(set, 10)

	if !reflect.DeepEqual(set, before) {
		t.Fatalf("the set changed: %v", set)
	}
}
