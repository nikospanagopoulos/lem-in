package paths

import (
	"errors"
	"testing"

	"lem-in/internal/graph"
)

// example03 reproduces the PRD's third audit example (section 6.5): a
// single path of length 3, start "0", end "5", 4 ants, expected 6 turns.
// Room "4" is declared before ##start in the original file, which is why
// it comes first here too — order inside Rooms must not matter to Sets.
func example03() graph.Colony {
	rooms := []graph.Room{
		{Name: "4", X: 5, Y: 4},
		{Name: "0", X: 1, Y: 4},
		{Name: "1", X: 3, Y: 6},
		{Name: "5", X: 6, Y: 4},
		{Name: "2", X: 3, Y: 4},
		{Name: "3", X: 3, Y: 1},
	}
	links := [][2]string{
		{"0", "1"},
		{"2", "4"},
		{"1", "4"},
		{"0", "2"},
		{"4", "5"},
		{"3", "0"},
		{"4", "3"},
	}
	return graph.Colony{
		Ants:  4,
		Start: "0",
		End:   "5",
		Rooms: rooms,
		Links: links,
	}
}

func TestSetsExample03(t *testing.T) {
	colony := example03()
	g := graph.New(colony)

	sets, err := Sets(g, colony.Start, colony.End)
	if err != nil {
		t.Fatalf("Sets returned error: %v", err)
	}
	if len(sets) == 0 {
		t.Fatal("Sets returned no sets at all")
	}

	// PRD table (6.5): example03 uses 1 path, total length 3.
	firstSet := sets[0]
	if len(firstSet) != 1 {
		t.Fatalf("sets[0] has %d paths, want 1", len(firstSet))
	}
	if got := firstSet[0].length(); got != 3 {
		t.Fatalf("path length = %d, want 3", got)
	}
}

func TestTurnsExample03(t *testing.T) {
	colony := example03()
	g := graph.New(colony)

	sets, err := Sets(g, colony.Start, colony.End)
	if err != nil {
		t.Fatalf("Sets returned error: %v", err)
	}

	best := Best(sets, colony.Ants)
	turns := Turns(best, colony.Ants)

	if turns != 6 {
		t.Fatalf("turns = %d, want 6 (PRD table, example03: 4 ants, 6 turns)", turns)
	}
}

func TestNoPath(t *testing.T) {
	colony := graph.Colony{
		Ants:  1,
		Start: "a",
		End:   "b",
		Rooms: []graph.Room{{Name: "a"}, {Name: "b"}},
		// no links: a and b are not connected at all
	}
	g := graph.New(colony)

	_, err := Sets(g, colony.Start, colony.End)
	if !errors.Is(err, ErrNoPath) {
		t.Fatalf("Sets returned %v, want ErrNoPath", err)
	}
}
