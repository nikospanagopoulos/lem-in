package sim

import "lem-in/internal/paths"

// Assign decides how many ants go down each path. The result has one number
// per path, in the same order as set: counts[i] is the number of ants that use
// set[i]. The numbers always add up to ants.
//
// The rule is the greedy rule of section 6.2 of the PRD. We take the ants one
// after the other, from ant 1 to the last ant, and each ant chooses the path
// where it would ARRIVE EARLIEST. If two paths give the same arrival turn, the
// ant takes the shorter one.
//
// How to compute the arrival turn of the next ant on a path:
//
//	arrival = length of the path + number of ants already sent down it
//
// For a path of 3 tunnels, the 1st ant arrives at turn 3 (3 + 0), the 2nd at
// turn 4 (3 + 1), and so on: ants enter a path one per turn, and each one needs
// "length" turns to cross.
//
// set must be sorted shortest path first. paths.Sets guarantees that (PRD 8.3),
// and we rely on it for the tie rule: see the comparison in the loop below.
//
// Assign never returns an error. A set with no path gives an empty result, and
// zero ants gives a result full of zeros. The parser already rejects a colony
// with no ants, so these two cases only protect the function from a misuse.
func Assign(set []paths.Path, ants int) []int {
	// counts starts as all zeros: nobody has been sent anywhere yet.
	counts := make([]int, len(set))
	if len(set) == 0 {
		return counts
	}

	// One round per ant. In each round we look at every path, find the
	// best one, and give that path one more ant. So the loop adds exactly 1
	// to exactly one counter per ant, and the counters add up to ants.
	for ant := 1; ant <= ants; ant++ {
		best := 0
		bestArrival := arrival(set[0], counts[0])

		for i := 1; i < len(set); i++ {
			a := arrival(set[i], counts[i])

			// Strictly "<", not "<=". On a tie we do NOT change our choice,
			// so we keep the path that comes first in the list. Because the
			// set is sorted shortest first, that is the shorter path. This
			// is the tie rule of the PRD, with no extra code for it.
			if a < bestArrival {
				best, bestArrival = i, a
			}
		}
		counts[best]++
	}
	return counts
}

// arrival is the turn at which the next ant sent down path p would reach the
// end room, when sent ants ants have already been sent down it.
//
// A path is a list of rooms, so its length in tunnels is one less than the
// number of rooms.
func arrival(p paths.Path, sent int) int {
	return len(p) - 1 + sent
}
