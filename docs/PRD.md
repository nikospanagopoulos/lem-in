# PRD — lem-in
*Team: dpavaloa (Coder A), dsakkas (Coder B), npanagopoul (Coder C) | Zone01 Athens*
*Scope: core project + specific error messages + ant farm visualizer (bonus)*

---

## 1. What We're Building and Why

`lem-in` is a Go program that reads the description of an ant colony from a file (rooms, tunnels, number of ants), moves every ant from `##start` to `##end` in the **fewest possible turns**, and prints the file followed by the moves, one line per turn.

It is the team's first project where the hard part is an **algorithm**, not plumbing: the colony is a graph, and the question "how do N ants cross it fastest without sharing a room" is a flow problem.

We ship:

- **Core**: parse, validate, solve, simulate, print
- **Specific error messages** (audit bonus)
- **Visualizer** (audit bonus): `./lem-in file.txt | ./visualizer`
- **Tests** for every package (audit bonus)

**Design goals, in priority order:**

1. **Correctness against the audit**: every example within its turn limit, every bad file rejected with an error, no crash on any input.
2. **Optimal, not "good enough"**: the turn count we print is the minimum possible, and we can explain why.
3. **Learnable architecture**: small packages with one job each, connected by contracts agreed before coding.
4. **Fast by design**: the algorithm's cost depends on the size of the colony, not on the number of ants.

Only the Go standard library is used.

---

## 2. Design Principles

- **One job per package.** Parsing does not know about paths. Path-finding does not print. Simulation does not read files.
- **Contracts first.** The exported types and function signatures in section 8 are agreed before anyone writes a function body. Changing a contract needs all three coders.
- **Never crash.** Every function that can fail returns an `error`. No `panic`, no `log.Fatal`, no `os.Exit` outside `main`.
- **Only `main` prints.** Every other function returns values. This is what makes each package testable alone.
- **Deterministic output.** Same file in, same output out, every run. No iteration over a Go `map` decides anything the user can see (map order is random).
- **Check by machine, not by eye.** Every output is replayed by a rule checker in the tests (section 12).
- **No concurrency.** The pipeline is strictly sequential (section 11).

---

## 3. Input Format

The file path is the only argument: `go run . file.txt`

| Line | Meaning |
|---|---|
| First line | Number of ants. Integer greater than 0. |
| `##start` | The next room line is the start room. |
| `##end` | The next room line is the end room. |
| `name x y` | A room. Three fields separated by single spaces. `x` and `y` are integers. |
| `name1-name2` | A tunnel between two rooms. Tunnels are two-way. |
| `#anything` | A comment. Ignored by the solver. |
| `##anything` (other than start/end) | Unknown command. Ignored by the solver. |

**Rules taken from the subject:**

- A room name never starts with `L` or `#` and contains no spaces.
- Room names are not necessarily numbers and are not in order.
- Rooms can appear before `##start` (see example03: room `4` is declared first).
- A tunnel joins exactly two rooms. Two rooms have at most one tunnel between them.

**Our own decisions (documented so all three coders parse the same way):**

- A room name cannot contain `-`, because a tunnel line is split on its single `-`.
- Rooms and tunnels are validated after the whole file is read, so their order in the file does not matter.
- A blank line inside the data is an error. Blank lines at the very end of the file are ignored.

---

## 4. Output Format

```
number_of_ants
the_rooms
the_links

Lx-y Lz-w Lr-o ...
```

1. **The input file, line by line, exactly as it was read**, including comments. (Audit example05 prints its `#rooms` comment in the expected output, so comments are echoed, not removed.)
2. One empty line.
3. One line per turn. Each move is `Lx-y`: ant number `x` moves into room `y`. Moves are separated by one space, **sorted by ant number**, with no trailing space.
4. Only ants that moved in that turn appear on its line.

**Example (subject, Example 1: 3 ants):**

```
3
##start
1 23 3
2 16 7
...
6-5

L1-3 L2-2
L1-4 L2-5 L3-3
L1-0 L2-6 L3-4
L2-0 L3-0
```

On error, the program prints one line and nothing else (section 9).

---

## 5. Movement Rules

- All ants start in `##start`. The run ends when all ants are in `##end`.
- An ant moves at most once per turn, through one tunnel.
- A tunnel is used at most once per turn.
- A room holds at most one ant. `##start` and `##end` hold any number.
- "The room must be empty" is judged on the turn as a whole: a room that an ant leaves during a turn can be entered by another ant in the same turn. Confirmed by audit example02, turn 2: `L2-2 L3-3 L4-1` (ant 2 leaves room `1`, ant 4 enters it).
- Ants are numbered from 1 to N. Lower numbers leave `##start` first.

---

## 6. The Algorithm

### 6.1 The problem, restated

Finding one shortest path is not enough. We must choose **a set of paths that share no room** (except start and end) and decide **how many ants go down each one**, so that the last ant arrives as early as possible.

Two facts shape everything:

- A path of length `L` (counted in tunnels) is a pipe: one ant can enter per turn, and each ant needs `L` turns to cross.
- More paths means more ants entering per turn, but each extra path is longer. Sometimes a second path helps, sometimes it does not. It depends on the number of ants.

### 6.2 How many turns a set of paths needs

Take paths with lengths `L1 ≤ L2 ≤ ... ≤ Lk`. The ant that enters path `i` in position `j` (1st, 2nd, ...) arrives at turn `Li + j − 1`.

**Rule for sending ants (greedy):** for each ant, from 1 to N, pick the path where it would arrive earliest, that is the path with the smallest `Li + (ants already assigned to it)`. On a tie, pick the shorter path.

**Turn count:** the largest arrival turn over all ants.

**Shortcut when every path receives at least one ant:**

```
turns = ceil( (ants + L1 + L2 + ... + Lk) / k ) − 1
```

Checks done by hand in our analysis:

| Case | Calculation | Turns |
|---|---|---|
| 3 ants, one path of length 3 | (3 + 3) / 1 − 1 | 5 |
| 3 ants, paths of length 3 and 4 | ceil((3 + 7) / 2) − 1 | 4 |

**Consequence:** for a fixed number of paths `k`, only the **sum of the lengths** matters. So the best set of `k` paths is the one with the smallest total length.

### 6.3 Chosen algorithm: successive shortest paths on a split-room network

This is minimum-cost flow with unit capacities (the idea behind Suurballe's algorithm). It finds, for k = 1, 2, 3, ..., the set of `k` room-disjoint paths with the **smallest total length**. We then keep the `k` that gives the fewest turns for our number of ants.

**Step 1: build the network.**
- Every room except start and end becomes two nodes, `in` and `out`, joined by an arc of capacity 1 and cost 0. This arc is what enforces "one path per room".
- Every tunnel `a-b` becomes two arcs: `a.out → b.in` and `b.out → a.in`, each with capacity 1 and cost 1.
- Every arc has a reverse arc with capacity 0 and the opposite cost. Sending flow along an arc adds capacity to its reverse.

**Step 2: find the cheapest route from start to end among arcs that still have capacity.**
Reverse arcs have cost −1, so plain BFS is not correct here. We use the queue-based Bellman-Ford search (a BFS that re-queues a node whenever a cheaper way to reach it is found).

**Step 3: push one unit of flow along that route.**
Travelling over a reverse arc **cancels** part of a path chosen earlier. This is how a later search can "reroute" an earlier path instead of being blocked by it.

**Step 4: record the result for this `k`.**
Read the `k` paths out of the network (follow the arcs that carry flow, from start, until end). Compute the turn count with the rule of 6.2.

**Step 5: repeat from step 2** until no route exists. The number of rounds is at most the number of tunnels touching start, or touching end, whichever is smaller.

**Step 6: keep the `k` with the fewest turns.** On a tie, keep the smaller `k`.

If step 2 fails on the very first round, there is no path between start and end: that is an error (section 9).

### 6.4 Why this one, and what we rejected

- **Successive shortest paths (min-cost flow)**: chosen. Optimal for every `k`, so optimal overall. Cost depends on the colony size only. Uses a queue and a graph search, which the team is learning anyway.
- ~~One BFS shortest path~~: rejected. On example01 it gives 13 turns (limit: 8).
- ~~Greedy: BFS, remove the rooms used, BFS again~~: rejected. On example01 the shortest path goes through `h` and `n`, which blocks the other routes. Result: 9 or 13 turns depending on tie-breaks (limit: 8). The correct answer needs three paths of length 5 that avoid the shortest path entirely.
- ~~Plain max-flow with BFS (Edmonds-Karp, no costs)~~: rejected. It finds the largest number of paths, but not the cheapest set for each `k`, so the turn count is not guaranteed minimal.
- ~~List every path and try every combination~~: rejected. The number of combinations explodes on colonies like example05.

### 6.5 Verified results

A throwaway prototype of 6.3 was run on every audit example, and each output was replayed by a rule checker (one ant per room, one use per tunnel per turn, all ants at end).

| Example | Ants | Paths used | Total length | Turns | Audit limit |
|---|---|---|---|---|---|
| Subject Example 1 | 3 | 2 | 7 | 4 | (shown: 4) |
| example00 | 4 | 1 | 3 | 6 | 6 |
| example01 | 10 | 3 | 15 | 8 | 8 |
| example02 | 20 | 2 | 4 | 11 | 11 |
| example03 | 4 | 1 | 3 | 6 | 6 |
| example04 | 9 | 2 | 5 | 6 | 6 |
| example05 | 9 | 3 | 17 | 8 | 8 |
| example06 | 100 | 2 | 5 | 52 | time only |
| example07 | 1000 | 2 | 5 | 502 | time only |

These numbers are the expected values for our tests.

---

## 7. Architecture Overview

One binary for the solver, one for the visualizer. One sequential pipeline:

```
main
  → parser.ParseFile      file      → Colony + original lines   (or error)
  → graph.New             Colony    → Graph
  → paths.Sets            Graph     → best set of k paths, for every k   (or ErrNoPath)
  → paths.Best            sets, ants → the set with the fewest turns
  → sim.Run               paths, ants → list of turns
  → output.Render         lines, turns → text
  → print
```

### 7.1 Repository structure

```
lem-in/
├── main.go                      # wiring only: run(args, w) error; no logic
├── main_test.go                 # end-to-end tests over testdata/
│
├── internal/
│   ├── graph/                   # Coder A: shared types + adjacency
│   │   ├── graph.go
│   │   └── graph_test.go
│   ├── parser/                  # Coder A: read + validate the file
│   │   ├── parser.go
│   │   ├── errors.go            # exported error values
│   │   └── parser_test.go
│   ├── paths/                   # Coder B: section 6.3
│   │   ├── network.go           # split-room network, arcs, reverse arcs
│   │   ├── paths.go             # Sets, Best, Turns
│   │   └── paths_test.go
│   ├── sim/                     # Coder C: who moves where, each turn
│   │   ├── sim.go
│   │   ├── check.go             # rule checker, used by tests
│   │   └── sim_test.go
│   └── output/                  # Coder C: final text
│       ├── output.go
│       └── output_test.go
│
├── cmd/
│   └── visualizer/              # bonus: reads lem-in output on stdin
│       └── main.go
│
├── testdata/                    # example00..07, badexample00..01, our own cases
├── docs/
│   ├── prd.md
│   └── task_cards.md
├── go.mod                       # module lem-in, no requires
├── .gitignore
└── README.md
```

`go run . file.txt` works from the repo root because `main.go` is there. `internal/` stops other modules importing our packages.

### 7.2 Dependencies

- `main` → `parser`, `graph`, `paths`, `sim`, `output`
- `parser` → `graph` (for the `Colony` type)
- `paths` → `graph`
- `sim` → `paths` (for the `Path` type)
- `output` → `sim` (for the `Turn` type)
- `cmd/visualizer` → `parser`, `graph`

No package imports `main`. No cycles.

---

## 8. Contracts: Shared Types and Exported Functions

Signatures only. Bodies are written by the owner. Unexported helpers are the owner's choice.

### 8.1 `graph` (Coder A)

```go
type Room struct {
    Name string
    X, Y int
}

type Colony struct {
    Ants  int
    Start string       // name of the start room
    End   string       // name of the end room
    Rooms []Room       // in file order
    Links [][2]string  // in file order
}

type Graph struct { /* unexported fields */ }

func New(c Colony) *Graph
func (g *Graph) Rooms() []string                 // all room names, file order
func (g *Graph) Neighbours(room string) []string // connected rooms, file order
```

### 8.2 `parser` (Coder A)

```go
func ParseFile(path string) (graph.Colony, []string, error)
func Parse(r io.Reader) (graph.Colony, []string, error) // same, testable without a file

var (
    ErrAnts          error // invalid number of Ants
    ErrNoStart       error // no start room found
    ErrNoEnd         error // no end room found
    ErrDuplicateRoom error // duplicated room
    ErrRoomFormat    error // invalid room name or coordinates
    ErrUnknownRoom   error // link to unknown room
    ErrSelfLink      error // room linked to itself
    ErrDuplicateLink error // duplicated link
    ErrLine          error // line is neither room, link, comment nor command
)
```

The `[]string` return value is the file's lines, unchanged, for `output.Render`.

### 8.3 `paths` (Coder B)

```go
type Path []string // room names; first is Start, last is End

var ErrNoPath error // no path between start and end

// Sets returns one path set per k: sets[0] is the best single path,
// sets[1] the best pair of room-disjoint paths, and so on.
func Sets(g *graph.Graph, start, end string) ([][]Path, error)

// Turns returns how many turns the set needs for this many ants (rule of 6.2).
func Turns(set []Path, ants int) int

// Best returns the set with the fewest turns. Tie: the smaller set.
func Best(sets [][]Path, ants int) []Path
```

Paths inside a set are sorted by length, shortest first.

### 8.4 `sim` (Coder C)

```go
type Move struct {
    Ant  int    // 1..N
    Room string // room the ant enters
}

type Turn []Move // sorted by Ant

// Assign returns how many ants go down each path (greedy rule of 6.2).
func Assign(set []paths.Path, ants int) []int

// Run returns every turn until all ants are in End.
func Run(set []paths.Path, ants int) []Turn

// Check replays turns against the rules of section 5. Used by tests.
func Check(c graph.Colony, turns []Turn) error
```

### 8.5 `output` (Coder C)

```go
func Render(lines []string, turns []sim.Turn) string
```

### 8.6 `main` (Coder C)

```go
func run(args []string, w io.Writer) error // main calls run(os.Args[1:], os.Stdout)
```

### 8.7 Naming and variable conventions

- Exported names (capital first letter) only for what another package needs. Everything else is unexported.
- No package-level variables that hold state. The only package-level variables are the `Err...` values.
- Receivers: `g` for `*Graph`. Loop variables: `room`, `link`, `path`, `ant`, `turn`, not `x`, `tmp`.
- A room is identified by its **name** in every contract. Using integer indexes inside a package for speed is allowed, as long as names come back out.
- `gofmt` formatting, `go vet` clean.

---

## 9. Error Handling

**Output on error:** one line, then the program returns normally.

```
ERROR: invalid data format
```

or, when we know the cause:

```
ERROR: invalid data format, <reason>
```

`main` does **not** call `os.Exit(1)`: with `go run .`, a non-zero exit adds an extra `exit status 1` line that is not in the audit's expected output.

| Case | Reason text |
|---|---|
| Ant count missing, not an integer, zero, negative, or too large for `int` | `invalid number of Ants` |
| No `##start`, or `##start` not followed by a room | `no start room found` |
| No `##end`, or `##end` not followed by a room | `no end room found` |
| Two `##start` or two `##end` | `duplicated start or end room` |
| Start and end are the same room | `start and end are the same room` |
| Two rooms with the same name | `duplicated room` |
| Room name starts with `L` or `#`, or contains `-` | `invalid room name` |
| Coordinate is not an integer | `invalid coordinates` |
| Tunnel names a room that does not exist | `link to unknown room` |
| Tunnel from a room to itself (`3-3`) | `room linked to itself` |
| Same tunnel twice (`7-8` and `8-7`) | `duplicated link` |
| Line that is not a room, tunnel, comment or command | `invalid line` |
| No path from start to end | `no path between start and end` |
| Wrong number of arguments, or file cannot be opened | `cannot read file` |

**Audit files:** `badexample00` has `0` ants. `badexample01` has a self-link (`3-3`), a duplicated link (`7-8`, `8-7`) and no tunnel reaching the end room. Any of these is enough to reject it.

**Implementation rules:**
- `parser` and `paths` return their exported `Err...` values. `main` turns them into the text above.
- No function panics on bad input. Indexing a slice or a map is always guarded.

---

## 10. Performance

### 10.1 Audit requirements

| Test | Limit |
|---|---|
| example06, 100 ants | under 1.5 minutes |
| example07, 1000 ants | under 2.5 minutes |

example06 and example07 are the six-room colony of example04 with more ants. They are small. The audit limits are easy; our own targets are stricter.

### 10.2 Team targets

| Test | Target |
|---|---|
| example06 (100 ants, 52 turns) | under 0.1 s |
| example07 (1000 ants, 502 turns) | under 0.5 s |
| Stress file of our own: 3000 rooms, 9000 tunnels, 1000 ants | under 2 s |
| Any audit example | under 0.1 s |

Measured with the built binary: `go build -o lem-in . && time ./lem-in testdata/example07.txt`

For reference, the Python prototype used for section 6.5 solved the stress case in 0.10 s, so these targets leave a wide margin for Go.

### 10.3 Where the time goes, and the rules that keep it low

- **Path-finding** costs (number of rounds) × (one graph search). Rounds are limited by the tunnels touching start or end. It does **not** depend on the number of ants.
- **Simulation and printing** cost about (ants × average path length) moves. For example07 that is about 2500 moves.

Rules:

1. Build the output with `strings.Builder` (or write through `bufio.Writer`). Never `s += ...` in a loop, never one `fmt.Println` per move.
2. Look up rooms through a `map[string]int` built once. Never search a slice for a room name inside a loop.
3. The simulation advances ants along their path by position. It never re-searches the graph during a turn.
4. `paths.Sets` runs once per program run. `paths.Turns` is arithmetic, not a simulation.
5. Measure before optimising. If a target is missed, find the slow function with `go test -bench` or `go tool pprof` first.

---

## 11. Concurrency

**Decision: no goroutines, channels or mutexes in `lem-in`.**

- Each stage needs the complete result of the stage before it. Nothing can run in parallel.
- The whole run takes milliseconds (section 10). There is nothing to gain.
- Shared state between goroutines would add data races to a project whose difficulty is already the algorithm.
- The audit asks for no concurrency.

The pipeline stays safe without locks because functions take arguments and return values, and no package keeps state.

**Visualizer:** a `time.Ticker` or `time.Sleep` for the animation delay is allowed. Still no shared state.

---

## 12. Testing

- **Unit tests per package**, table-driven, in `*_test.go` next to the code.
- **`testdata/`** holds the eight audit examples, the two bad examples, and our own cases.
- **Rule checker (`sim.Check`)**: replays a list of turns and fails if an ant uses a missing tunnel, a tunnel is used twice in a turn, two ants end a turn in the same room, an ant moves twice, or an ant is not in `##end` at the finish.
- **End-to-end test (`main_test.go`)**: for every valid example, run the pipeline, pass the result to `sim.Check`, and compare the number of turns to the table in section 6.5.

**Cases each owner must cover:**

| Package | Cases |
|---|---|
| `parser` | every row of the error table in section 9; rooms declared before `##start`; comments; unknown `##command`; names that are words; trailing blank line |
| `graph` | neighbours of start, end, a middle room, an isolated room |
| `paths` | example01 returns three paths of total length 15; example03 returns one path; direct start-end tunnel (example02); no path returns `ErrNoPath`; `Turns` for the two rows of 6.2 |
| `sim` | 3 ants on paths 3 and 4 gives 4 turns; 1 ant; ants fewer than paths; path of length 1 (start-end); every result passes `Check` |
| `output` | comments echoed; exactly one blank line; no trailing spaces; moves sorted by ant |

---

## 13. Visualizer (Bonus)

- **Usage:** `./lem-in file.txt | ./visualizer`
- **Input:** the solver's output on stdin. The first block (up to the blank line) is the colony, parsed with `parser.Parse`. The second block is the turns.
- **Display:** in the terminal. Rooms are placed on a character grid from their coordinates, tunnels are drawn between them, and each ant is shown in its room. One frame per turn, redrawn with ANSI escape codes. Standard library only.
- **Controls:** a delay between frames, set by a flag (for example `-delay 300ms`).
- **Errors:** if stdin starts with `ERROR`, print it and stop.
- **Built last**, after every core acceptance criterion is green.

Audit points targeted: "presents an ant farm visualizer" and "possible to see the ants moving". The 3D visualizer point is not targeted.

---

## 14. How We Split the Work

| Part | Owner | Reviewer |
|---|---|---|
| `graph`: shared types, adjacency | dpavaloa | dsakkas |
| `parser`: reading, validation, error values | dpavaloa | dsakkas |
| `paths`: network, search, `Sets`, `Turns`, `Best` | dsakkas | npanagopoul |
| `sim`: `Assign`, `Run`, `Check` | npanagopoul | dpavaloa |
| `output` and `main` | npanagopoul | dpavaloa |
| `testdata/` and end-to-end test | npanagopoul | dsakkas |
| Visualizer | TEAM (owner to be chosen) | TEAM |
| README | TEAM | TEAM |

Reviewer ≠ author. A branch is merged only after its reviewer approves.

**What each coder learns:**

- **Coder A (dpavaloa):** reading files with `bufio.Scanner`, string handling, validation, designing error values, structs and maps, representing a graph.
- **Coder B (dsakkas):** graphs, queues, BFS and its Bellman-Ford extension, flow networks, residual arcs, proving a result is optimal.
- **Coder C (npanagopoul):** modelling time step by step, greedy assignment, writing a checker, wiring packages together, output formatting, end-to-end testing, measuring performance.

**Order of work:**

1. Coder A pushes the types of 8.1 (no logic) on day one. Everyone builds on them.
2. All three write their package against the contracts, with hand-written inputs in tests. Nobody waits for anybody.
3. Coder C wires `main` when `ParseFile`, `Sets` and `Run` exist.
4. End-to-end tests, then performance checks, then the visualizer.

---

## 15. Acceptance Criteria

### Functional (audit)
- [ ] `go run . example00.txt` reads the colony and prints the expected format
- [ ] Only `##start` and `##end` are treated as commands; other `##` lines are ignored
- [ ] One line per turn, moves as `Lx-y`, sorted by ant
- [ ] example00 ≤ 6 turns, example01 ≤ 8, example02 ≤ 11, example03 ≤ 6, example04 ≤ 6, example05 ≤ 8
- [ ] badexample00 and badexample01 print `ERROR: invalid data format` (with or without a reason)
- [ ] example06 (100 ants) and example07 (1000 ants) meet the targets of section 10.2
- [ ] One ant per room, one use per tunnel per turn (proved by `sim.Check`)
- [ ] All ants are in `##end` at the finish
- [ ] Same output on repeated runs
- [ ] Every row of the error table (section 9) prints its message; no input causes a panic

### Bonus (audit)
- [ ] Specific error messages
- [ ] Visualizer shows the colony and the ants moving
- [ ] Test file for every package; tests cover the cases of section 12
- [ ] Good practices: `gofmt -l .` empty, `go vet ./...` clean, no dead code, no global state

### Code quality
- [ ] Only standard packages (`go.mod` has no `require`)
- [ ] `go build ./...` and `go test ./...` pass
- [ ] Every exported name of section 8 exists with the agreed signature
- [ ] Only `main` and the visualizer print

### Team
- [ ] Each coder can explain their own package and the contract with its neighbours without the code in front of them
- [ ] Each coder can explain, on example01, why the shortest path is not part of the answer

---

## 16. Out of Scope

- 3D visualizer
- Graphical window or web interface
- Reading the colony from stdin in `lem-in` (the subject passes a file argument)
- Input formats other than the subject's
- Concurrency
- External packages

---

## 17. Open Decisions

- **Visualizer owner:** one coder, or pair-programmed by all three after the core is done.
- **Room names containing `-`:** rejected for now (section 3). Revisit only if an auditor's file needs them.

---

*Last updated: 2026-11-12 | Team: dpavaloa, dsakkas, npanagopoul*