package paths

import (
	"errors"
	"testing"

	"lem-in/internal/graph"
)

// --- Colony builders -------------------------------------------------
//
// Each builder reproduces one audit example exactly as given in the PRD's
// source documents (the Functional audit doc), room order included, so
// Sets sees the same file order a real run would. Room order matters for
// determinism (PRD section 2), which is why these are not simplified.

func example00Colony() graph.Colony {
	return graph.Colony{
		Ants:  4,
		Start: "0",
		End:   "1",
		Rooms: []graph.Room{
			{Name: "0", X: 0, Y: 3},
			{Name: "2", X: 2, Y: 5},
			{Name: "3", X: 4, Y: 0},
			{Name: "1", X: 8, Y: 3},
		},
		Links: [][2]string{
			{"0", "2"}, {"2", "3"}, {"3", "1"},
		},
	}
}

func example01Colony() graph.Colony {
	return graph.Colony{
		Ants:  10,
		Start: "start",
		End:   "end",
		Rooms: []graph.Room{
			{Name: "start", X: 1, Y: 6},
			{Name: "0", X: 4, Y: 8},
			{Name: "o", X: 6, Y: 8},
			{Name: "n", X: 6, Y: 6},
			{Name: "e", X: 8, Y: 4},
			{Name: "t", X: 1, Y: 9},
			{Name: "E", X: 5, Y: 9},
			{Name: "a", X: 8, Y: 9},
			{Name: "m", X: 8, Y: 6},
			{Name: "h", X: 4, Y: 6},
			{Name: "A", X: 5, Y: 2},
			{Name: "c", X: 8, Y: 1},
			{Name: "k", X: 11, Y: 2},
			{Name: "end", X: 11, Y: 6},
		},
		Links: [][2]string{
			{"start", "t"}, {"n", "e"}, {"a", "m"}, {"A", "c"},
			{"0", "o"}, {"E", "a"}, {"k", "end"}, {"start", "h"},
			{"o", "n"}, {"m", "end"}, {"t", "E"}, {"start", "0"},
			{"h", "A"}, {"e", "end"}, {"c", "k"}, {"n", "m"}, {"h", "n"},
		},
	}
}

func example02Colony() graph.Colony {
	return graph.Colony{
		Ants:  20,
		Start: "0",
		End:   "3",
		Rooms: []graph.Room{
			{Name: "0", X: 2, Y: 0},
			{Name: "1", X: 4, Y: 1},
			{Name: "2", X: 6, Y: 0},
			{Name: "3", X: 5, Y: 3},
		},
		Links: [][2]string{
			{"0", "1"}, {"0", "3"}, {"1", "2"}, {"3", "2"},
		},
	}
}

func example03Colony() graph.Colony {
	return graph.Colony{
		Ants:  4,
		Start: "0",
		End:   "5",
		Rooms: []graph.Room{
			{Name: "4", X: 5, Y: 4},
			{Name: "0", X: 1, Y: 4},
			{Name: "1", X: 3, Y: 6},
			{Name: "5", X: 6, Y: 4},
			{Name: "2", X: 3, Y: 4},
			{Name: "3", X: 3, Y: 1},
		},
		Links: [][2]string{
			{"0", "1"}, {"2", "4"}, {"1", "4"}, {"0", "2"},
			{"4", "5"}, {"3", "0"}, {"4", "3"},
		},
	}
}

func example04Colony() graph.Colony {
	return graph.Colony{
		Ants:  9,
		Start: "richard",
		End:   "peter",
		Rooms: []graph.Room{
			{Name: "richard", X: 0, Y: 6},
			{Name: "gilfoyle", X: 6, Y: 3},
			{Name: "erlich", X: 9, Y: 6},
			{Name: "dinish", X: 6, Y: 9},
			{Name: "jimYoung", X: 11, Y: 7},
			{Name: "peter", X: 14, Y: 6},
		},
		Links: [][2]string{
			{"richard", "dinish"}, {"dinish", "jimYoung"},
			{"richard", "gilfoyle"}, {"gilfoyle", "peter"},
			{"gilfoyle", "erlich"}, {"richard", "erlich"},
			{"erlich", "jimYoung"}, {"jimYoung", "peter"},
		},
	}
}

// --- Helpers -----------------------------------------------------------

func sumLengths(set []Path) int {
	total := 0
	for _, p := range set {
		total += p.length()
	}
	return total
}

func containsRoom(p Path, room string) bool {
	for _, r := range p {
		if r == room {
			return true
		}
	}
	return false
}

// forwardArcTo scans the arcs leaving node "from" for the one real
// (non-reverse) arc that goes to "to". Reverse arcs always have origCap 0,
// so origCap > 0 is what tells them apart (see network.go, addArc).
func forwardArcTo(net *network, from, to int) (arc, bool) {
	for _, a := range net.arcs[from] {
		if a.origCap > 0 && a.to == to {
			return a, true
		}
	}
	return arc{}, false
}

// --- LEM-12 / LEM-13: total length and turns per k, against the PRD's
// own table (section 6.5), for every audit example whose shape is small
// enough to write by hand. example01 is covered separately below
// (TestTrapExample01), because it also needs the "trap" assertion.
func TestSetsAndTurnsAuditExamples(t *testing.T) {
	cases := []struct {
		name            string
		colony          graph.Colony
		wantPaths       int // k chosen by Best
		wantTotalLength int // sum of lengths in that set
		wantTurns       int
	}{
		{"example00", example00Colony(), 1, 3, 6},
		{"example02_direct_tunnel", example02Colony(), 2, 4, 11},
		{"example03", example03Colony(), 1, 3, 6},
		{"example04", example04Colony(), 2, 5, 6},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := graph.New(tc.colony)

			sets, err := Sets(g, tc.colony.Start, tc.colony.End)
			if err != nil {
				t.Fatalf("Sets returned error: %v", err)
			}

			best := Best(sets, tc.colony.Ants)
			if len(best) != tc.wantPaths {
				t.Errorf("Best returned %d paths, want %d", len(best), tc.wantPaths)
			}
			if got := sumLengths(best); got != tc.wantTotalLength {
				t.Errorf("total length = %d, want %d", got, tc.wantTotalLength)
			}

			turns := Turns(best, tc.colony.Ants)
			if turns != tc.wantTurns {
				t.Errorf("turns = %d, want %d", turns, tc.wantTurns)
			}
		})
	}
}

// --- LEM-10: the "trap" case. The single shortest path in example01 goes
// through rooms h and n (PRD section 6.4), and greedily locking that path
// in first is exactly what breaks a naive BFS-and-remove approach: it
// blocks the three length-5 paths that the real optimum needs. Sets must
// still land on k=3, total length 15, turns 8 (PRD table, section 6.5) —
// which it can only do if later searches reroute around what the first
// search chose, i.e. by using reverse arcs (PRD 6.3, step 3).
func TestTrapExample01UsesReverseArc(t *testing.T) {
	colony := example01Colony()
	g := graph.New(colony)

	sets, err := Sets(g, colony.Start, colony.End)
	if err != nil {
		t.Fatalf("Sets returned error: %v", err)
	}

	// The trap: the very first path found (k=1) is the short one through
	// h and n.
	if len(sets[0]) != 1 {
		t.Fatalf("sets[0] has %d paths, want 1", len(sets[0]))
	}
	trap := sets[0][0]
	if trap.length() != 4 {
		t.Errorf("k=1 path length = %d, want 4 (the start-h-n-e-end trap)", trap.length())
	}
	if !containsRoom(trap, "h") || !containsRoom(trap, "n") {
		t.Errorf("k=1 path %v does not go through h and n as expected", trap)
	}

	// The escape: Best still finds the global optimum, which needs the
	// paths that the trap would have blocked.
	best := Best(sets, colony.Ants)
	if len(best) != 3 {
		t.Fatalf("Best chose %d paths, want 3", len(best))
	}
	if got := sumLengths(best); got != 15 {
		t.Errorf("total length of best set = %d, want 15", got)
	}
	if got := Turns(best, colony.Ants); got != 8 {
		t.Errorf("turns = %d, want 8", got)
	}
}

// --- LEM-09: exact network structure on a 3-room chain, and proof that
// building the network never mutates the graph it was built from.
func TestNetworkChainStructure(t *testing.T) {
	colony := graph.Colony{
		Ants:  1,
		Start: "start",
		End:   "end",
		Rooms: []graph.Room{
			{Name: "start"}, {Name: "A"}, {Name: "end"},
		},
		Links: [][2]string{
			{"start", "A"}, {"A", "end"},
		},
	}
	g := graph.New(colony)

	// Snapshot the graph before building the network, to check afterwards
	// that newNetwork only reads it.
	beforeRooms := g.Rooms()
	beforeNeighbours := map[string][]string{}
	for _, room := range beforeRooms {
		beforeNeighbours[room] = g.Neighbours(room)
	}

	net, err := newNetwork(g, "start", "end")
	if err != nil {
		t.Fatalf("newNetwork returned error: %v", err)
	}

	// 3 rooms * 2 nodes each.
	if got := len(net.arcs); got != 6 {
		t.Fatalf("network has %d nodes, want 6", got)
	}

	startID, endID := 0, 2 // file order: start, A, end
	aID := 1

	// in->out arcs: start and end are unlimited, A is capacity 1.
	if a, ok := forwardArcTo(net, nodeIn(startID), nodeOut(startID)); !ok || a.cap != unlimited || a.cost != 0 {
		t.Errorf("start in->out arc = %+v, ok=%v, want cap=unlimited cost=0", a, ok)
	}
	if a, ok := forwardArcTo(net, nodeIn(aID), nodeOut(aID)); !ok || a.cap != 1 || a.cost != 0 {
		t.Errorf("A in->out arc = %+v, ok=%v, want cap=1 cost=0", a, ok)
	}
	if a, ok := forwardArcTo(net, nodeIn(endID), nodeOut(endID)); !ok || a.cap != unlimited || a.cost != 0 {
		t.Errorf("end in->out arc = %+v, ok=%v, want cap=unlimited cost=0", a, ok)
	}

	// Tunnel arcs: start-A, both directions.
	if a, ok := forwardArcTo(net, nodeOut(startID), nodeIn(aID)); !ok || a.cap != 1 || a.cost != 1 {
		t.Errorf("start->A tunnel arc = %+v, ok=%v, want cap=1 cost=1", a, ok)
	}
	if a, ok := forwardArcTo(net, nodeOut(aID), nodeIn(startID)); !ok || a.cap != 1 || a.cost != 1 {
		t.Errorf("A->start tunnel arc = %+v, ok=%v, want cap=1 cost=1", a, ok)
	}

	// Tunnel arcs: A-end, both directions.
	if a, ok := forwardArcTo(net, nodeOut(aID), nodeIn(endID)); !ok || a.cap != 1 || a.cost != 1 {
		t.Errorf("A->end tunnel arc = %+v, ok=%v, want cap=1 cost=1", a, ok)
	}
	if a, ok := forwardArcTo(net, nodeOut(endID), nodeIn(aID)); !ok || a.cap != 1 || a.cost != 1 {
		t.Errorf("end->A tunnel arc = %+v, ok=%v, want cap=1 cost=1", a, ok)
	}

	// 3 in->out arcs + 4 tunnel-direction arcs = 7 forward arcs, each with
	// a reverse arc, so 14 arc entries in total across the whole network.
	total := 0
	for _, arcs := range net.arcs {
		total += len(arcs)
	}
	if total != 14 {
		t.Errorf("total arc entries = %d, want 14", total)
	}

	// The graph g must be exactly as it was before newNetwork ran.
	afterRooms := g.Rooms()
	if len(afterRooms) != len(beforeRooms) {
		t.Fatalf("g.Rooms() changed length: before %v, after %v", beforeRooms, afterRooms)
	}
	for i, room := range afterRooms {
		if room != beforeRooms[i] {
			t.Errorf("g.Rooms()[%d] changed: before %q, after %q", i, beforeRooms[i], room)
		}
		if got := g.Neighbours(room); !equalSlices(got, beforeNeighbours[room]) {
			t.Errorf("g.Neighbours(%q) changed: before %v, after %v", room, beforeNeighbours[room], got)
		}
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// --- No route at all: start and end are not connected, no panic.
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

// --- LEM-13: Best with a genuine tie, and Turns with fewer ants than
// paths. A 4-room diamond gives two room-disjoint paths of equal length,
// so Best must apply the "smaller k wins" tie-break rule (PRD 8.3) between
// k=1 (one of the two paths alone) and k=2 (both together) only when their
// turn counts actually tie — which happens with very few ants.
func TestBestTieBreakAndFewAnts(t *testing.T) {
	colony := graph.Colony{
		Ants:  1,
		Start: "s",
		End:   "e",
		Rooms: []graph.Room{
			{Name: "s"}, {Name: "a"}, {Name: "b"}, {Name: "e"},
		},
		Links: [][2]string{
			{"s", "a"}, {"a", "e"}, {"s", "b"}, {"b", "e"},
		},
	}
	g := graph.New(colony)

	sets, err := Sets(g, colony.Start, colony.End)
	if err != nil {
		t.Fatalf("Sets returned error: %v", err)
	}
	if len(sets) != 2 {
		t.Fatalf("got %d sets, want 2 (k=1 and k=2)", len(sets))
	}

	// With only 1 ant, one path alone and two paths together both finish
	// in the same number of turns (the single ant only ever uses one of
	// them) — a genuine tie, so Best must prefer the smaller set.
	best := Best(sets, 1)
	if len(best) != 1 {
		t.Errorf("Best(sets, 1) returned %d paths, want 1 (tie-break to smaller k)", len(best))
	}

	// With plenty of ants, two parallel paths win outright: no tie.
	best = Best(sets, 10)
	if len(best) != 2 {
		t.Errorf("Best(sets, 10) returned %d paths, want 2", len(best))
	}

	// Fewer ants than paths: Turns must still return a sane, small value,
	// not something built for assuming every path gets at least one ant.
	turns := Turns(sets[1], 1)
	if turns != 2 {
		t.Errorf("Turns(2 paths of length 2, 1 ant) = %d, want 2", turns)
	}
}
