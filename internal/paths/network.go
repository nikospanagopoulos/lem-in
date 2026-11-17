// network.go implements the split-room flow network described in the PRD
// (section 6.3): every room becomes two nodes joined by a capacity-1 arc,
// tunnels become capacity-1 arcs between rooms, and every arc has a reverse
// arc that lets a later search "undo" part of an earlier one.
//
// Node numbering (integer-indexed, not string-keyed, for speed — PRD
// section 10.3): each room gets an integer id i (0, 1, 2, ...). Node 2*i is
// its "in" side, node 2*i+1 is its "out" side. This is why every lookup in
// this file is a slice index, never a map lookup, on the hot path.
package paths

import (
	"math"

	"lem-in/internal/graph"
)

// unreachable marks a node the current search has not reached yet. It must
// be larger than any real distance the network can produce.
const unreachable = math.MaxInt32

// noNode marks "no node" in the prev-node / prev-arc trails used to
// reconstruct a path after a search.
const noNode = -1

// unlimited is the capacity given to the in->out arc of the start and end
// rooms. The subject allows any number of ants in those two rooms, so this
// arc must never be the bottleneck. It is a large int, not literally
// infinite — no real colony pushes this much flow.
const unlimited = math.MaxInt32 / 2

// arc is one directed, capacitated, costed edge of the residual network.
// Every arc has a paired reverse arc (see addArc) that starts at capacity 0
// and gains capacity as flow is pushed along the forward arc — this is what
// lets a later search cancel part of an earlier one.
type arc struct {
	to      int // destination node
	cap     int // remaining capacity on THIS arc, right now
	origCap int
	cost    int // cost of sending one unit of flow along this arc
	rev     int // index of the reverse arc, inside arcs[to]
}

// network is the split-room graph of one colony, built once by newNetwork
// and then searched repeatedly (once per k) by shortestPath / augment.
type network struct {
	arcs      [][]arc        // arcs[node] = the arcs leaving that node
	roomIndex map[string]int // room name -> room id (0..len(roomNames)-1)
	roomNames []string       // room id -> room name (the reverse lookup)
	start     int            // node id of the start room's "in" side
	end       int            // node id of the end room's "in" side
}

// nodeIn and nodeOut turn a room id into its two node ids. Kept as named
// functions, not inlined everywhere, so the 2*i / 2*i+1 relationship is
// written down in exactly one place.
func nodeIn(room int) int  { return 2 * room }
func nodeOut(room int) int { return 2*room + 1 }

// newNetwork builds the split-room network of g. start and end must be room
// names that exist in g; the caller (paths.Sets) gets them from a
// graph.Colony the parser has already validated, so this function does not
// re-check them.
func newNetwork(g *graph.Graph, start, end string) (*network, error) {
	roomNames := g.Rooms()

	roomIndex := make(map[string]int, len(roomNames))
	for i, name := range roomNames {
		roomIndex[name] = i
	}

	startID, ok := roomIndex[start]
	if !ok {
		return nil, ErrNoPath
	}
	endID, ok := roomIndex[end]
	if !ok {
		return nil, ErrNoPath
	}
	net := &network{
		arcs:      make([][]arc, 2*len(roomNames)),
		roomIndex: roomIndex,
		roomNames: roomNames,
		start:     nodeIn(startID),
		end:       nodeIn(endID),
	}

	// One in->out arc per room. Capacity 1 enforces "one path per room";
	// start and end are exempt, because any number of ants may wait there.
	for i := range roomNames {
		capacity := 1
		if i == startID || i == endID {
			capacity = unlimited
		}
		net.addArc(nodeIn(i), nodeOut(i), capacity, 0)
	}

	// One arc per tunnel, in the direction room -> neighbour. Ranging over
	// every room's neighbours covers both directions of every tunnel,
	// because "a-b" makes b a neighbour of a AND a a neighbour of b.
	for i, name := range roomNames {
		for _, neighbour := range g.Neighbours(name) {
			net.addArc(nodeOut(i), nodeIn(roomIndex[neighbour]), 1, 1)
		}
	}

	return net, nil
}

// addArc adds a forward arc from->to and its paired reverse arc to->from.
// The reverse starts at capacity 0: it carries no flow until the forward arc
// does, and then it can carry back exactly as much as was sent forward.
func (net *network) addArc(from, to, capacity, cost int) {
	net.arcs[from] = append(net.arcs[from], arc{
		to:      to,
		cap:     capacity,
		origCap: capacity,
		cost:    cost,
		rev:     len(net.arcs[to]), // where the reverse arc will land, below
	})
	net.arcs[to] = append(net.arcs[to], arc{
		to:      from,
		cap:     0,
		origCap: 0,
		cost:    -cost,
		rev:     len(net.arcs[from]) - 1, // where the forward arc just landed, above
	})
}

// step is one node of a reconstructed path: the node itself, and the index
// (inside arcs[the PREVIOUS node]) of the arc that was taken to reach it.
// step.arc is noNode for the path's first element (the start), which was
// not reached by any arc.
type step struct {
	node int
	arc  int
}

// shortestPath finds the cheapest route from start to end using only arcs
// that still have capacity. Reverse arcs have negative cost, so this cannot
// be a plain BFS or Dijkstra: it uses the queue-based Bellman-Ford
// relaxation of PRD section 6.3 (often called SPFA). found is false when no
// route exists at all — the caller turns that into paths.ErrNoPath.
func (net *network) shortestPath() (path []step, cost int, found bool) {
	dist := make([]int, len(net.arcs))
	prevNode := make([]int, len(net.arcs))
	prevArc := make([]int, len(net.arcs))
	queued := make([]bool, len(net.arcs))
	for i := range dist {
		dist[i] = unreachable
		prevNode[i] = noNode
		prevArc[i] = noNode
	}

	dist[net.start] = 0
	queue := []int{net.start}
	queued[net.start] = true

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		queued[node] = false

		for idx, a := range net.arcs[node] {
			if a.cap <= 0 {
				continue // no residual capacity: this arc cannot be used right now
			}
			candidate := dist[node] + a.cost
			if candidate < dist[a.to] {
				dist[a.to] = candidate
				prevNode[a.to] = node
				prevArc[a.to] = idx
				if !queued[a.to] {
					queue = append(queue, a.to)
					queued[a.to] = true
				}
			}
		}
	}

	if dist[net.end] == unreachable {
		return nil, 0, false
	}

	// Walk the prev trail back from end to start, then reverse it, so the
	// returned path reads start -> ... -> end.
	for node := net.end; ; node = prevNode[node] {
		path = append(path, step{node: node, arc: prevArc[node]})
		if node == net.start {
			break
		}
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path, dist[net.end], true
}

// augment sends one unit of flow along path: every forward arc on the path
// loses one unit of capacity, and its reverse arc gains one unit. A later
// shortestPath call can then use that reverse arc to undo this choice, which
// is how the algorithm reconsiders an earlier path (PRD 6.3, step 3).
func (net *network) augment(path []step) {
	for i := 1; i < len(path); i++ {
		from := path[i-1].node
		idx := path[i].arc
		a := &net.arcs[from][idx]
		a.cap--
		net.arcs[a.to][a.rev].cap++
	}
}

// roomPath turns a node path into the room names it passes through, in
// order, with no repeats: a room contributes its name once, at its "in"
// node — every path visits each room's "in" node exactly once, so the "out"
// node is skipped to avoid listing the room twice.
func (net *network) roomPath(path []step) []string {
	rooms := make([]string, 0, len(path)/2+1)
	for _, s := range path {
		if s.node%2 == 0 { // an "in" node: see nodeIn/nodeOut above
			rooms = append(rooms, net.roomNames[s.node/2])
		}
	}
	return rooms
}

// flowRemaining counts, for every "real" arc (one with origCap > 0 — a
// reverse arc always has origCap 0 and is skipped), how many units of flow
// currently pass through it. That is origCap - cap: capacity only ever
// moves from a forward arc to its reverse, so this difference is exactly
// what was sent and not since cancelled.
func (net *network) flowRemaining() [][]int {
	remaining := make([][]int, len(net.arcs))
	for node, arcs := range net.arcs {
		remaining[node] = make([]int, len(arcs))
		for idx, a := range arcs {
			if a.origCap > 0 {
				remaining[node][idx] = a.origCap - a.cap
			}
		}
	}
	return remaining
}

// decomposeOnePath walks one path from start to end, following any arc
// that still has unconsumed flow, and consumes it as it goes. Because every
// room (other than start/end) allows at most one unit of flow through its
// in->out arc, this walk can never revisit a room: it is always a simple
// path.
func (net *network) decomposeOnePath(remaining [][]int) []step {
	path := []step{{node: net.start, arc: noNode}}
	node := net.start
	for node != net.end {
		for idx := range net.arcs[node] {
			if remaining[node][idx] > 0 {
				remaining[node][idx]--
				node = net.arcs[node][idx].to
				path = append(path, step{node: node, arc: idx})
				break
			}
		}
	}
	return path
}

// decomposePaths reads k room-disjoint paths out of the network's current
// flow (PRD 6.3, step 4). This is run once, after k rounds of
// shortestPath+augment have pushed k units of flow in total.
func (net *network) decomposePaths(k int) []Path {
	remaining := net.flowRemaining()
	paths := make([]Path, 0, k)
	for i := 0; i < k; i++ {
		steps := net.decomposeOnePath(remaining)
		paths = append(paths, Path(net.roomPath(steps)))
	}
	return paths
}
