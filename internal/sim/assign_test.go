package sim

import (
	"math/rand"
	"slices"
	"testing"

	"lem-in/internal/paths"
)

// pathOf builds a path with the given number of tunnels. Assign only looks at
// the length of a path, so the names of the rooms do not matter here.
func pathOf(tunnels int) paths.Path {
	return make(paths.Path, tunnels+1)
}

// setOf builds a set of paths from lengths, written shortest first, like a set
// that comes out of paths.Sets.
func setOf(lengths ...int) []paths.Path {
	set := make([]paths.Path, len(lengths))
	for i, l := range lengths {
		set[i] = pathOf(l)
	}
	return set
}

// The cases of the task card, plus the hand calculation for 5 ants on paths of
// 2 and 4 tunnels, plus the edge cases.
func TestAssign(t *testing.T) {
	tests := []struct {
		name    string
		lengths []int
		ants    int
		want    []int
	}{
		// Card: 3 ants on paths of 3 and 4 tunnels give [2, 1].
		{"3 ants, paths of 3 and 4", []int{3, 4}, 3, []int{2, 1}},
		// Card: 1 ant, 2 paths: the shorter path gets it.
		{"1 ant goes down the shorter path", []int{1, 10}, 1, []int{1, 0}},
		// Worked by hand: ants 1,2 go on the short path (arrivals 2, 3),
		// ant 3 ties (4 and 4) and takes the short path, ant 4 takes the
		// long path (5 against 4), ant 5 ties (5 and 5) and takes the short.
		{"5 ants, paths of 2 and 4", []int{2, 4}, 5, []int{4, 1}},
		// Tie rule: equal paths, one ant. It takes the first one.
		{"a tie goes to the first path", []int{2, 2}, 1, []int{1, 0}},
		// Equal paths share the ants as evenly as possible.
		{"equal paths share the ants", []int{2, 2}, 4, []int{2, 2}},
		{"one path takes every ant", []int{5}, 7, []int{7}},
		{"three paths", []int{1, 3, 5}, 6, []int{4, 2, 0}},
		{"zero ants", []int{2, 4}, 0, []int{0, 0}},
		{"no paths", nil, 5, []int{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Assign(setOf(tc.lengths...), tc.ants)
			if !slices.Equal(got, tc.want) {
				t.Fatalf("Assign(%v, %d) = %v, want %v", tc.lengths, tc.ants, got, tc.want)
			}
		})
	}
}

// turnsOf computes how many turns a given assignment takes: the last ant on
// each path that is used arrives at turn (length + count - 1).
func turnsOf(lengths, counts []int) int {
	last := 0
	for i, count := range counts {
		if count > 0 {
			last = max(last, lengths[i]+count-1)
		}
	}
	return last
}

// randomLengths returns k path lengths, sorted shortest first.
func randomLengths(rng *rand.Rand, k int) []int {
	lengths := make([]int, k)
	for i := range lengths {
		lengths[i] = 1 + rng.Intn(15)
	}
	slices.Sort(lengths)
	return lengths
}

// For any set and any number of ants, the counts must add up to the number of
// ants, and no count can be negative. We try many random cases, because one
// hand-made case could hide a mistake that only appears in another shape.
func TestAssignCountsAddUpToAnts(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 20000; i++ {
		lengths := randomLengths(rng, 1+rng.Intn(6))
		ants := rng.Intn(60)

		counts := Assign(setOf(lengths...), ants)

		sum := 0
		for _, c := range counts {
			if c < 0 {
				t.Fatalf("lengths %v, ants %d: negative count in %v", lengths, ants, counts)
			}
			sum += c
		}
		if sum != ants {
			t.Fatalf("lengths %v, ants %d: counts %v add up to %d", lengths, ants, counts, sum)
		}
	}
}

// Assign and paths.Turns are written by different people with different
// methods: ant by ant here, and a binary search there. If they ever disagree,
// one of them is wrong. This is the consistency test of the card (the check
// against Run comes with Run, in the next card).
func TestAssignAgreesWithPathsTurns(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	for i := 0; i < 20000; i++ {
		lengths := randomLengths(rng, 1+rng.Intn(6))
		ants := 1 + rng.Intn(60)
		set := setOf(lengths...)

		got := turnsOf(lengths, Assign(set, ants))
		want := paths.Turns(set, ants)
		if got != want {
			t.Fatalf("lengths %v, ants %d: Assign gives %d turns, paths.Turns gives %d",
				lengths, ants, got, want)
		}
	}
}

// Assign must only read the set it receives.
func TestAssignDoesNotChangeTheSet(t *testing.T) {
	set := []paths.Path{{"s", "a", "e"}, {"s", "b", "c", "e"}}
	before := []paths.Path{{"s", "a", "e"}, {"s", "b", "c", "e"}}

	Assign(set, 10)

	for i := range set {
		if !slices.Equal(set[i], before[i]) {
			t.Fatalf("path %d changed: %v", i, set[i])
		}
	}
}
