// Package paths finds, for every k = 1, 2, 3, ..., the set of k
// room-disjoint paths from start to end with the smallest total length
// (PRD section 6.3), and picks the k that gives the fewest turns for a
// given number of ants (PRD section 6.2).
package paths

import (
	"errors"
	"math"
	"sort"

	"lem-in/internal/graph"
)

// ErrNoPath means start and end are not connected at all: the very first
// search in Sets found no route.
var ErrNoPath = errors.New("no path between start and end")

// Path is a sequence of room names: Path[0] is Start, Path[len(Path)-1] is
// End.
type Path []string

// length is the path's length in tunnels, which is one less than its
// number of rooms.
func (p Path) length() int { return len(p) - 1 }

// Sets returns one path set per k: sets[0] is the best single path,
// sets[1] the best pair of room-disjoint paths, and so on, up to the
// largest k for which such a set exists. Paths inside each set are sorted
// by length, shortest first.
func Sets(g *graph.Graph, start, end string) ([][]Path, error) {
	net := newNetwork(g, start, end)

	var sets [][]Path
	for k := 1; ; k++ {
		route, _, found := net.shortestPath()
		if !found {
			break // no more disjoint routes exist at all
		}
		net.augment(route)

		paths := net.decomposePaths(k)
		sort.Slice(paths, func(i, j int) bool {
			return paths[i].length() < paths[j].length()
		})
		sets = append(sets, paths)
	}

	if len(sets) == 0 {
		return nil, ErrNoPath
	}
	return sets, nil
}

// Turns returns how many turns this set of paths needs to move ants
// of them across, under the greedy assignment rule of PRD section 6.2:
// each ant goes down whichever path lets it arrive earliest.
//
// Rather than simulate that assignment ant by ant (which would cost time
// proportional to the number of ants — too slow, and a problem in its own
// right for a huge ant count), this does a binary search on the answer:
// for a candidate turn count T, a path of length L can get
// max(0, T-L+1) ants there by turn T (its first ant arrives at turn L, the
// next at L+1, and so on). Summing that over every path in the set gives
// the most ants this set can deliver by turn T. The smallest T for which
// that sum reaches the ant count is the answer — and it is reached by the
// same greedy assignment the PRD describes, so this gives the identical
// turn count without ever touching an individual ant.
func Turns(set []Path, ants int) int {
	if len(set) == 0 || ants <= 0 {
		return 0
	}

	lengths := make([]int, len(set))
	minLen := math.MaxInt
	for i, p := range set {
		lengths[i] = p.length()
		if lengths[i] < minLen {
			minLen = lengths[i]
		}
	}

	// lo is never enough (0 ants fit the formula at T < minLen), hi is
	// always enough: a single path alone, used by every ant one after
	// another, finishes by minLen + ants - 1, so minLen + ants always
	// covers it even with only one path in the set.
	lo, hi := minLen, minLen+ants
	for lo < hi {
		mid := lo + (hi-lo)/2
		if deliverable(lengths, mid) >= ants {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo
}

// deliverable is how many ants a set of paths (given by their lengths) can
// get to the end by turn t, if every path is kept as full as possible.
func deliverable(lengths []int, t int) int {
	total := 0
	for _, l := range lengths {
		if room := t - l + 1; room > 0 {
			total += room
		}
	}
	return total
}

// Best returns the set from sets that needs the fewest turns for ants
// ants. On a tie, the smaller set wins (PRD section 8.3).
func Best(sets [][]Path, ants int) []Path {
	best := sets[0]
	bestTurns := Turns(best, ants)

	for _, set := range sets[1:] {
		turns := Turns(set, ants)
		if turns < bestTurns || (turns == bestTurns && len(set) < len(best)) {
			best, bestTurns = set, turns
		}
	}
	return best
}
