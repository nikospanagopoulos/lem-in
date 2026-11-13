# Task Cards — lem-in
*Team: dpavaloa (Coder A), dsakkas (Coder B), npanagopoul (Coder C) | Zone01 Athens*
*Based on: `prd.md` v1.0 · Contracts: PRD section 8 · Repo: `platform.zone01.gr/git/npanagopoul/lem-in`*

---

## 1. How to Use These Cards

**One card = one branch = one pull request.** Cards are small enough to finish and review in one sitting.

**Workflow for every card**

1. **Claim it.** Set the status on the board (section 3) to `In progress`.
2. **Pseudocode first.** Before writing Go, write a short note (10–20 lines, in the PR description) answering the card's *Think about first* questions. The reviewer reads it and asks questions. Only then you write code. This is where the learning happens.
3. **Branch:** `feat/lem-NN-short-name`. Never push to `main` directly.
4. **Code and tests together.** Tests listed on the card are part of the card, not a follow-up.
5. **Run `./check.sh`** (created in LEM-01) before opening the PR.
6. **Review.** The reviewer on the card approves. The reviewer is never the author. The author must be able to explain the code without looking at it.
7. **Merge, then tick the box** on the board.

**Size guide (rough):** S under 1 h · M 1–3 h · L 3 h or more.

**Status legend:** `[ ]` To do · `[~]` In progress · `[r]` In review · `[x]` Done

### Definition of Done (applies to every card)

- [ ] Names and signatures match PRD section 8 exactly
- [ ] `gofmt -l .` prints nothing, `go vet ./...` is clean, `go build ./...` passes
- [ ] Tests for the cases on the card are written and pass (`go test ./...`)
- [ ] Every exported name has a doc comment
- [ ] No `panic`, no `log.Fatal`, no `os.Exit` outside `main`; errors are returned
- [ ] No package-level state (only `Err...` values)
- [ ] Only standard packages imported
- [ ] Pseudocode note exists in the PR
- [ ] Reviewer approved

### Reviewer checklist

- Does it do what the card says, and nothing else?
- Can I break it with an input the author did not think of? (Try empty, one item, duplicates, huge.)
- Does anything depend on Go `map` iteration order?
- Is each error returned, not swallowed?
- Could I explain this function to a third person after reading it?

---

## 2. Delivery Waves

| Wave | Goal | Cards |
|---|---|---|
| 0 | Foundations: everyone can build and import | LEM-01, 02, 03 |
| 1 | Packages in parallel, each tested alone | LEM-04 to 17 |
| 2 | Wire together, prove, measure | LEM-18 to 20 |
| 3 | Stretch and bonus | LEM-21 to 24 |
| 4 | Finish | LEM-25, 26 |

LEM-02 (shared types) is the single blocker. Coder A pushes it first, within the first hour.

---

## 3. Board

| ID | Card | Owner | Reviewer | Depends on | Size | Status |
|---|---|---|---|---|---|---|
| LEM-01 | Repo skeleton and tooling | npanagopoul | dsakkas | none | S | `[ ]` |
| LEM-02 | Shared types in `graph` | dpavaloa | dsakkas | 01 | S | `[ ]` |
| LEM-03 | Test data: audit and own colonies | npanagopoul | dsakkas | 01 | S | `[ ]` |
| LEM-04 | Parser: reader, raw lines, ant count | dpavaloa | dsakkas | 02 | M | `[ ]` |
| LEM-05 | Parser: rooms and commands | dpavaloa | dsakkas | 04 | M | `[ ]` |
| LEM-06 | Parser: tunnels and final validation | dpavaloa | dsakkas | 05 | M | `[ ]` |
| LEM-07 | Parser: error table and robustness tests | dpavaloa | dsakkas | 06 | M | `[ ]` |
| LEM-08 | `graph.New`, `Rooms`, `Neighbours` | dpavaloa | dsakkas | 02 | S | `[ ]` |
| LEM-09 | Paths: split-room network | dsakkas | npanagopoul | 02 | M | `[ ]` |
| LEM-10 | Paths: cheapest augmenting route | dsakkas | npanagopoul | 09 | L | `[ ]` |
| LEM-11 | Paths: push flow, read paths out | dsakkas | npanagopoul | 10 | M | `[ ]` |
| LEM-12 | Paths: `Sets` | dsakkas | npanagopoul | 08, 11 | M | `[ ]` |
| LEM-13 | Paths: `Turns` and `Best` | dsakkas | npanagopoul | 02 | M | `[ ]` |
| LEM-14 | Sim: `Assign` | npanagopoul | dpavaloa | 13 | S | `[ ]` |
| LEM-15 | Sim: `Run` | npanagopoul | dpavaloa | 14 | L | `[ ]` |
| LEM-16 | Sim: `Check` (independent rule checker) | npanagopoul | dpavaloa | 02 | M | `[ ]` |
| LEM-17 | Output: `Render` | npanagopoul | dpavaloa | 15 | S | `[ ]` |
| LEM-18 | `main`: `run` and error messages | npanagopoul | dpavaloa | 07, 12, 15, 17 | M | `[ ]` |
| LEM-19 | End-to-end tests over `testdata/` | npanagopoul | dsakkas | 03, 18 | M | `[ ]` |
| LEM-20 | Stress file and performance check | dpavaloa | dsakkas | 18 | M | `[ ]` |
| LEM-21 | Stretch: optimality cross-check | dsakkas + npanagopoul | dpavaloa | 19 | L | `[ ]` |
| LEM-22 | Visualizer: read the solver output | dpavaloa *(proposed)* | npanagopoul | 18 | M | `[ ]` |
| LEM-23 | Visualizer: layout rooms and tunnels | dpavaloa *(proposed)* | dsakkas | 22 | L | `[ ]` |
| LEM-24 | Visualizer: animate the turns | npanagopoul *(proposed)* | dpavaloa | 23 | M | `[ ]` |
| LEM-25 | README | TEAM | TEAM | 19 | M | `[ ]` |
| LEM-26 | Audit dry run | TEAM | TEAM | all | M | `[ ]` |

The visualizer owners are a proposal: PRD section 17 leaves the choice open. Coder B carries the hardest cards, so B is not given visualizer work.

---

## 4. Wave 0: Foundations

### LEM-01 · Repo skeleton and tooling
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| npanagopoul | dsakkas | none | S |

**Goal:** a repo where all three can clone, build and import each other's packages from minute one.

**Deliverables**
- All three members added as collaborators on the Gitea repo
- `go.mod` with `module lem-in` and no `require` lines
- The folder layout of PRD section 7.1, each package with a `package` file that compiles empty
- `.gitignore` for built binaries (`lem-in`, `visualizer`)
- `check.sh`: runs `gofmt -l .`, `go vet ./...`, `go build ./...`, `go test ./...` and fails if any of them does
- `docs/prd.md` and `docs/task_cards.md` committed
- Short rule in `README.md`: no direct pushes to `main`, one PR per card

**Think about first**
- What does `internal/` stop? Who can import `lem-in/internal/graph`?
- Why must `main.go` live in the repo root for `go run . file.txt` to work?
- What is the difference between `go build ./...` and `go vet ./...`? What does each catch that the other misses?

**Acceptance criteria**
- [ ] A fresh `git clone` followed by `./check.sh` passes
- [ ] `go run . ` with no arguments compiles and runs (it may print an error line)
- [ ] `go.mod` has no `require` block

**You will learn:** Go modules, package layout, a review-based git workflow.

---

### LEM-02 · Shared types in `graph`
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| dpavaloa | dsakkas | LEM-01 | S |

**Goal:** unblock everyone. Push the types and signatures before any logic exists.

**Deliverables**
- `Room`, `Colony`, `Graph` and the signatures of `New`, `Rooms`, `Neighbours` exactly as in PRD 8.1. Bodies may be stubs that return zero values.
- A doc comment on every exported name saying what it holds and who produces it.

**Think about first**
- `Colony.Rooms` is a slice, not a map. What would change in the program's output if it were a map? (Think about printing and about repeatable runs.)
- Which lookups will `paths` need on a `Graph`? Neighbours by name, all rooms. Does it need anything else?
- Why does every contract identify a room by its name, while a package is still free to use integers inside?

**Acceptance criteria**
- [ ] Compiles; `paths`, `sim` and `parser` can import it
- [ ] Signatures match PRD 8.1 character for character
- [ ] Pushed and announced to the team on day one

**You will learn:** designing a contract before the implementation, exported vs unexported names.

---

### LEM-03 · Test data: audit and own colonies
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| npanagopoul | dsakkas | LEM-01 | S |

**Goal:** one shared folder of input files that every test uses.

**Deliverables**
- `testdata/example00.txt` to `example07.txt`, and `badexample00.txt`, `badexample01.txt`, copied exactly from the audit examples page
- Own small cases, one file each: a direct start-to-end tunnel; one ant; fewer ants than paths; no path; comments and an unknown `##command`; rooms declared before `##start`; room names that are words
- `testdata/README.md`: a table `file → ants → expected best turns → audit limit`, using PRD section 6.5

**Think about first**
- Which rule of the subject does each of your own cases attack?
- The audit's expected output for example02 contains a typo (`L16 -1`). What does that tell you about comparing the text of an output versus its number of turns?

**Acceptance criteria**
- [ ] All ten audit files present and byte-identical to the audit page
- [ ] At least seven own cases, each with a one-line comment in the README saying why it exists
- [ ] Expected turns match PRD 6.5

**You will learn:** reading a specification as a list of testable claims.

---

## 5. Wave 1A: Parser and Graph (Coder A)

### LEM-04 · Parser: reader, raw lines, ant count
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| dpavaloa | dsakkas | LEM-02 | M |

**Goal:** read a file line by line, keep the original text, and validate line 1.

**Deliverables**
- `Parse(r io.Reader)` and `ParseFile(path)` with the signatures of PRD 8.2. `ParseFile` opens the file and calls `Parse`.
- The returned `[]string` holds the file's lines **unchanged**, except: trailing blank lines are dropped and a trailing `\r` is removed from each line.
- `ErrAnts`, returned for a missing, non-integer, zero, negative or overflowing ant count.

**Think about first**
- Why does `Parse` take an `io.Reader` instead of a path? What does that let you do in a test?
- What does `bufio.Scanner` do with a very long line? What happens to the rest of the file?
- Which of these are valid ant counts: `3`, `+3`, ` 3`, `3 `, `03`, `0`, `-1`, `99999999999999999999`, `three`, an empty first line? Decide each one and write it down.
- Why do we keep the raw lines at all? Who uses them?

**Acceptance criteria**
- [ ] Table-driven test over the ant-count cases above
- [ ] An empty file and a file with only blank lines return an error, never a panic
- [ ] Raw lines returned equal the file, minus trailing blank lines

**You will learn:** `bufio.Scanner`, `io.Reader`, `strconv`, designing for testability.

---

### LEM-05 · Parser: rooms and commands
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| dpavaloa | dsakkas | LEM-04 | M |

**Goal:** recognise rooms, `##start`, `##end`, comments and unknown commands.

**Deliverables**
- A room line `name x y` produces a `Room`; coordinates are integers (negative allowed).
- `##start` / `##end` mark the **next room line**; any other `##...` line and any `#...` line is ignored.
- Errors: `ErrNoStart`, `ErrNoEnd`, `ErrDuplicateRoom`, `ErrRoomFormat`, plus a duplicated start/end and "start equals end".
- Room-name rules: not starting with `L` or `#`, no spaces, and (our decision, PRD 3) no `-`.

**Think about first**
- To know that "the next room is the start", what does the parser have to remember between two lines? Draw the states.
- What if `##start` is followed by a comment? By a tunnel? By another `##start`? By the end of the file?
- How do you tell a room line from a tunnel line? From a comment?
- A room is called `L`, `Lx`, `#a`, `a b`, `a-b`: which are legal under the subject and which under our own rules?

**Acceptance criteria**
- [ ] Rooms declared before `##start` are accepted (example03)
- [ ] `##start` or `##end` with no room after it returns the matching error
- [ ] Two rooms with the same name return `ErrDuplicateRoom`
- [ ] Unknown `##command` lines are ignored, not rejected

**You will learn:** a small state machine, `strings.Fields`, validating input defensively.

---

### LEM-06 · Parser: tunnels and final validation
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| dpavaloa | dsakkas | LEM-05 | M |

**Goal:** read tunnels and validate the whole colony once the file is finished.

**Deliverables**
- A tunnel line `a-b` produces a `[2]string` link.
- After the last line: both ends of every tunnel exist; no self-link; no duplicated tunnel (`7-8` and `8-7` are the same tunnel); start and end both exist.
- Errors: `ErrUnknownRoom`, `ErrSelfLink`, `ErrDuplicateLink`, `ErrLine` for any line that is nothing else.

**Think about first**
- Why validate the ends of a tunnel after the whole file is read, not as the line arrives?
- How do you detect that `7-8` and `8-7` are the same tunnel? What key would you store in a set?
- If a room name could contain `-`, why would `a-b-c` be ambiguous? How does our rule remove the problem?
- What if a tunnel line appears before some rooms? Does your parser accept it? Is that what you intend?

**Acceptance criteria**
- [ ] badexample01 is rejected (self-link, duplicate tunnel, unknown tunnel end)
- [ ] Link order and room order in the file do not change the result
- [ ] The returned `Colony.Links` keeps file order

**You will learn:** sets with `map`, normalising a key, validating across a whole structure.

---

### LEM-07 · Parser: error table and robustness tests
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| dpavaloa | dsakkas | LEM-06 | M |

**Goal:** prove that every row of the PRD error table is detected and that nothing can crash the parser.

**Deliverables**
- One test case per row of PRD section 9, each checked with `errors.Is`
- A robustness test: empty file, only newlines, only comments, a binary blob, a 1 MB line, a line of 10 000 `-`, Windows line endings
- A Go fuzz test `FuzzParse` (standard `testing` package) that asserts "never panics"

**Think about first**
- Why check with `errors.Is` and not by comparing text?
- Which inputs could make an index out of range, a nil map write or an infinite loop?
- What does the fuzzer do that your hand-written cases do not?

**Acceptance criteria**
- [ ] Every row of PRD 9 has a test that names the row in its description
- [ ] `go test -run Fuzz` and a short `go test -fuzz=FuzzParse -fuzztime=30s` finish with no crash
- [ ] Parser statement coverage reported in the PR (`go test -cover`)

**You will learn:** table-driven tests, `errors.Is`, fuzzing.

---

### LEM-08 · `graph.New`, `Rooms`, `Neighbours`
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| dpavaloa | dsakkas | LEM-02 | S |

**Goal:** the adjacency structure that `paths` and the visualizer read.

**Deliverables**
- `New(Colony) *Graph` building adjacency in both directions
- `Rooms()` in file order; `Neighbours(room)` in file order of the tunnels

**Think about first**
- A tunnel is two-way. Where do you record that?
- What does `Neighbours` return for a room that does not exist? For an isolated room?
- If the caller modifies the returned slice, what happens to the graph? How do you stop that? (PRD 8: a function does not change its inputs; the same care applies to what it returns.)

**Acceptance criteria**
- [ ] Tests: start room, end room, a middle room, an isolated room, an unknown room
- [ ] Order of neighbours is the same on every run
- [ ] Modifying a returned slice does not change later calls

**You will learn:** adjacency lists, protecting internal state, deterministic order.

---

## 6. Wave 1B: Paths (Coder B)

Read PRD section 6 first, especially 6.2 and 6.3. These five cards build the algorithm in the order it runs.

### LEM-09 · Paths: split-room network
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| dsakkas | npanagopoul | LEM-02 | M |

**Goal:** turn a `Graph` into the flow network of PRD 6.3, step 1. Internal to the package (`network.go`), not exported.

**Deliverables**
- Node numbering for the `in` and `out` copy of each room
- Arcs with capacity and cost, each paired with its reverse arc
- A way to go from a node back to its room name

**Think about first**
- Why does one room become two nodes? What does the arc between them, with capacity 1, enforce?
- Why do start and end need an unlimited capacity inside them?
- A tunnel is two-way: why does it become two arcs, and what does each arc's capacity 1 mean?
- Why does every arc have a partner with capacity 0 and cost minus 1? What does it represent?
- How do you store an arc and its partner so that updating one updates the other in constant time?

**Acceptance criteria**
- [ ] Test on a 3-room chain: the exact number of nodes and arcs, and the capacities and costs
- [ ] Test on example00: the network has the expected size
- [ ] Building the network does not change the `Graph`

**You will learn:** modelling a constraint (one ant per room) as part of a graph.

---

### LEM-10 · Paths: cheapest augmenting route
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| dsakkas | npanagopoul | LEM-09 | L |

**Goal:** PRD 6.3, step 2. Find the cheapest route from start to end through arcs that still have capacity.

**Deliverables**
- A function returning the route as a list of arcs, or "no route"
- The queue-based Bellman-Ford search described in PRD 6.3

**Think about first**
- Plain BFS finds the route with the fewest arcs. Once reverse arcs of cost minus 1 exist, why is that no longer the cheapest route?
- Take example01. The shortest path goes through `h` and `n`. Why can a later route need to *undo* part of it? Draw it.
- When you re-queue a node because you found a cheaper way to it, what stops the search from running forever? (Think about whether a negative cycle can exist here, and why not.)
- What do you keep per node so that you can rebuild the route afterwards?
- What is the worst-case cost, in terms of nodes and arcs?

**Acceptance criteria**
- [ ] On a graph with no route, it reports "no route" without error or panic
- [ ] On the "trap" graph you draw yourself (shortest path blocks two others), the second search uses a reverse arc
- [ ] Deterministic: same route on every run

**You will learn:** shortest paths with negative costs, why greedy fails, residual graphs.

---

### LEM-11 · Paths: push flow, read paths out
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| dsakkas | npanagopoul | LEM-10 | M |

**Goal:** PRD 6.3, steps 3 and 4. Apply a route, then read the current set of paths out of the network.

**Deliverables**
- Push one unit of flow along a route, updating arcs and their partners
- Read out `k` paths as room names (start to end), each room once, shortest path first, ties in a fixed order

**Think about first**
- After a push, how many ants-worth of capacity does a used arc have left? And its partner?
- How do you turn "arcs that carry flow" back into a list of room names? Where do `in` and `out` nodes collapse into one name?
- When two rounds have cancelled part of an earlier path, does the final set still contain the same paths as before? Do you rebuild from scratch each round or follow the flow?
- Could unused flow form a loop that belongs to no path? Why or why not? (All tunnels cost 1.)

**Acceptance criteria**
- [ ] Paths in a set share no room except start and end
- [ ] Every path starts at start and ends at end
- [ ] Order is the same on every run

**You will learn:** reading a result out of a residual network, determinism.

---

### LEM-12 · Paths: `Sets`
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| dsakkas | npanagopoul | LEM-08, LEM-11 | M |

**Goal:** the exported `Sets(g, start, end) ([][]Path, error)` of PRD 8.3.

**Deliverables**
- Rounds of search-and-push until there is no route; after each round, the paths for that `k` are stored (`sets[k-1]` holds `k` paths)
- `ErrNoPath` if the first round finds nothing

**Think about first**
- How many rounds can there be at most? Name the two quantities that bound it.
- Does round `k+1` keep the paths of round `k`, or may it change them? Which is true, and how does your test show it?
- What does a caller need if a colony has only one possible path?

**Acceptance criteria** (verified in the prototype of PRD 6.5; total length is counted in tunnels)

| Colony | Total length of the paths for k = 1, 2, 3, 4 |
|---|---|
| subject Example 1 | 3, 7 |
| example00 | 3 |
| example01 | 4, 9, 15 |
| example02 | 1, 4 |
| example03 | 3 |
| example04 | 2, 5 |
| example05 | 4, 10, 17, 25 |

- [ ] The table above holds in a test
- [ ] No path to the end returns `ErrNoPath`
- [ ] Start directly linked to end (example02) works

**You will learn:** assembling a result from rounds, testing against known answers.

---

### LEM-13 · Paths: `Turns` and `Best`
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| dsakkas | npanagopoul | LEM-02 | M |

**Goal:** how many turns a set of paths needs, and which set to choose. Pure arithmetic: no simulation.

**Deliverables**
- `Turns(set, ants)`: exact, following the greedy rule of PRD 6.2
- `Best(sets, ants)`: the set with the fewest turns; on a tie, the smaller set

**Think about first**
- Derive the turn count for `k` paths when every path receives an ant. Check it against 3 ants on paths of 3 and 4 tunnels (answer in the PRD).
- Now take 1 ant and paths of lengths 1 and 10. Is the shortcut still right? What must `Turns` do instead? Why is "exact" better than "a formula that is usually right"?
- `sim` imports `paths`, so `paths` cannot call `sim`. What does that force you to do about the greedy rule, and how will you make sure `sim.Assign` and `Turns` never disagree? (LEM-14 owns the consistency test.)
- Why is the tie-break "smaller set"? Look at example05, `k = 3` versus `k = 4`.

**Acceptance criteria** (expected turns for each `k`)

| Colony (ants) | k = 1 | k = 2 | k = 3 | k = 4 | Best |
|---|---|---|---|---|---|
| subject Example 1 (3) | 5 | 4 | | | 4 |
| example00 (4) | 6 | | | | 6 |
| example01 (10) | 13 | 9 | 8 | | 8 |
| example02 (20) | 20 | 11 | | | 11 |
| example03 (4) | 6 | | | | 6 |
| example04 (9) | 10 | 6 | | | 6 |
| example05 (9) | 12 | 9 | 8 | 8 | 8, with 3 paths |

- [ ] The table above holds in a test
- [ ] A test with fewer ants than paths
- [ ] A test with a tie

**You will learn:** reasoning about a pipeline, why exact beats approximate.

---

## 7. Wave 1C: Simulation and Output (Coder C)

### LEM-14 · Sim: `Assign`
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| npanagopoul | dpavaloa | LEM-13 | S |

**Goal:** `Assign(set, ants) []int`: how many ants go down each path.

**Deliverables**
- The greedy rule of PRD 6.2: each ant, in order, takes the path where it would arrive earliest; ties go to the shorter path
- A **consistency test** between `Assign`, `Run` and `paths.Turns`: for many (set, ants) pairs, the number of turns from `Run` equals `Turns`

**Think about first**
- Why is "the path where this ant arrives earliest" the right choice for every ant, one after the other?
- Ants are numbered 1 to N and lower numbers leave first. Which ants end up on which path?
- Paths are sorted shortest first. How does that give you the tie-break for free?

**Acceptance criteria**
- [ ] 3 ants on paths of 3 and 4 tunnels: counts `[2, 1]`
- [ ] 1 ant, 2 paths: the shorter path gets it
- [ ] Sum of the counts always equals the number of ants
- [ ] The consistency test passes on the audit colonies and on random sets

**You will learn:** greedy choice, a test that ties two packages together.

---

### LEM-15 · Sim: `Run`
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| npanagopoul | dpavaloa | LEM-14 | L |

**Goal:** `Run(set, ants) []Turn`: every turn until all ants are in the end room.

**Deliverables**
- Each turn: ants already inside a path advance one room, then each path releases its next waiting ant into its first room
- Moves of a turn sorted by ant number
- Ant numbering consistent with `Assign` (lower numbers leave first)

**Think about first**
- What do you need to remember about each ant? (Which path, which position.)
- Why can two ants on the same path never collide? Why can two ants on different paths never collide? (Which PRD property of the set do you rely on?)
- Why can a tunnel never be used twice in one turn here? Do you have to check, or does the model guarantee it?
- A path of length 1 (start linked directly to end): how many ants can use it per turn? Why?
- Can a turn ever be empty before the last ant arrives?
- How do you detect the end of the run without counting turns in advance?

**Acceptance criteria**
- [ ] 3 ants on paths of 3 and 4 tunnels: exactly 4 turns, and the turns match the hand-written ones of the PRD
- [ ] Example02: 11 turns; example01: 8 turns; example05: 8 turns
- [ ] Number of turns equals `paths.Turns` (consistency test of LEM-14)
- [ ] Every result passes `Check` (LEM-16)
- [ ] No empty turn line

**You will learn:** time-step simulation, invariants you can rely on instead of checking.

---

### LEM-16 · Sim: `Check` (independent rule checker)
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| npanagopoul | dpavaloa | LEM-02 | M |

**Goal:** `Check(colony, turns) error`: replay a list of turns against the rules of PRD section 5, **without reusing any logic from `Run`** and without knowing the paths. It sees what an auditor sees: the colony and the moves.

**Deliverables**
- Errors that name the turn number, the ant and the rule broken
- Rules: an ant moves at most once per turn; every move uses an existing tunnel from the ant's current room; a tunnel is used at most once per turn; at the end of a turn no room except start and end holds two ants; an ant in the end room never leaves it; at the finish every ant is in the end room

**Think about first**
- Why must `Check` be written independently of `Run`? What would you learn from a checker that shares code with the thing it checks?
- The rule "a room that empties during a turn can be entered in the same turn": do you check occupancy before the moves, after, or both? What does example02, turn 2 tell you?
- How do you compute where each ant is at the start of each turn? What do you store?

**Acceptance criteria**
- [ ] One handwritten *bad* turn list per rule above, each returning an error that names the rule
- [ ] The valid outputs of the subject (Examples 1, 2, 3) pass
- [ ] Reviewer (dpavaloa) tries to produce a violation `Check` misses

**You will learn:** why independent verification matters, writing negative tests.

---

### LEM-17 · Output: `Render`
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| npanagopoul | dpavaloa | LEM-15 | S |

**Goal:** `Render(lines, turns) string`: the exact text of PRD section 4.

**Deliverables**
- The raw lines unchanged, one empty line, then one line per turn of moves `L<ant>-<room>` separated by a single space, sorted by ant, with no trailing space

**Think about first**
- Why `strings.Builder` and not `s += ...` in a loop? What does the second cost for 1000 ants and 500 turns?
- What does your output end with: a newline or not? What would the audit terminal show either way?
- Why must `Render` not compute anything?

**Acceptance criteria**
- [ ] Test: Example 1 of the subject produces the text of the subject character for character
- [ ] Comments in the input appear in the output (example05)
- [ ] No trailing space on any line; exactly one empty line before the turns

**You will learn:** exact formatting, `strings.Builder`.

---

## 8. Wave 2: Integration

### LEM-18 · `main`: `run` and error messages
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| npanagopoul | dpavaloa | LEM-07, LEM-12, LEM-15, LEM-17 | M |

**Goal:** wire the pipeline and turn every error into the right line.

**Deliverables**
- `run(args []string, w io.Writer) error`: arguments check, `ParseFile`, `graph.New`, `Sets`, `Best`, `Run`, `Render`
- `main` calls `run` and, on error, prints **one line**: `ERROR: invalid data format` plus `, <reason>` from PRD section 9; then it returns normally (no `os.Exit(1)`)
- Mapping from every `Err...` value to its reason text; an unknown error gives the plain message

**Think about first**
- With `go run .`, what extra line does a non-zero exit status add? Why is that a problem for the audit?
- `errors.Is` or `==` when an error has been wrapped with `%w`? Why does wrapping matter for the reason text?
- Where does a *wrong number of arguments* or an unreadable file belong in the table?
- What must `main` not contain? (PRD 2: no logic.)

**Acceptance criteria**
- [ ] `go run . example00.txt` prints a valid result
- [ ] badexample00 and badexample01 print exactly one `ERROR:` line and nothing else
- [ ] No argument, a missing file, a directory as argument: one `ERROR:` line, no panic
- [ ] Only `main` prints

**You will learn:** wiring packages, error wrapping, why exit codes matter to tools.

---

### LEM-19 · End-to-end tests over `testdata/`
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| npanagopoul | dsakkas | LEM-03, LEM-18 | M |

**Goal:** one test file that runs the whole program on every file and proves the result.

**Deliverables**
- `main_test.go`: for every valid file, run the pipeline, parse the moves back, call `Check` and compare the number of turns to `testdata/README.md`
- Every bad file produces the expected `ERROR:` line
- Determinism: two runs on the same file give identical bytes

**Think about first**
- Why do you parse the moves back from the output text instead of testing the internal turn list?
- Which assertion catches a solver that returns a valid but *too slow* answer? Which catches an *invalid* one?

**Acceptance criteria**
- [ ] All audit examples within their limits (PRD 6.5 table, final column)
- [ ] The two bad examples and all own bad cases rejected
- [ ] Determinism test passes

**You will learn:** end-to-end testing, closing the loop on the audit.

---

### LEM-20 · Stress file and performance check
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| dpavaloa | dsakkas | LEM-18 | M |

**Goal:** measure against the team targets of PRD 10.2.

**Deliverables**
- A test helper that generates, with a fixed seed, a colony of 3000 rooms, 9000 tunnels and 1000 ants (generated at test time, not committed)
- Benchmarks: example06, example07, the stress colony
- A line in `README.md` with the measured times

**Think about first**
- Why does the number of ants barely change the path-finding time but does change the printing time?
- Before you optimise, what would you measure? (`go test -bench`, `go tool pprof`.)
- Is a random colony guaranteed to have a path from start to end? What does your generator do about it?

**Acceptance criteria**
- [ ] example06 under 0.1 s, example07 under 0.5 s, stress colony under 2 s
- [ ] The stress result passes `Check`
- [ ] If a target is missed: a profile and a fix, with before/after numbers in the PR

**You will learn:** benchmarking, profiling, reading a cost model.

---

## 9. Wave 3: Stretch and Bonus

### LEM-21 · Stretch: optimality cross-check
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| dsakkas + npanagopoul | dpavaloa | LEM-19 | L |

**Goal:** show on small colonies that our turn count is the true minimum, not just good on the audit examples.

**Deliverables**
- A brute-force search for tiny colonies (at most about 6 rooms and 4 ants) that tries every legal move sequence and returns the fewest turns
- A test that generates many random tiny colonies and compares the brute-force answer to the solver's

**Think about first**
- What is the *state* of the colony at one moment? How big is the space of states for 6 rooms and 4 ants?
- Which search finds the fewest turns? (You met it in the first lesson.)
- If the two answers ever differ, which of them is wrong, and how would you find out?

**Acceptance criteria**
- [ ] At least 500 random tiny colonies compared, seed printed on failure
- [ ] Any mismatch is either fixed or documented with a minimal example

**You will learn:** how to test an algorithm against an oracle.

---

### LEM-22 · Visualizer: read the solver output
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| dpavaloa *(proposed)* | npanagopoul | LEM-18 | M |

**Goal:** `./lem-in file | ./visualizer` starts by understanding its input.

**Deliverables**
- `cmd/visualizer`: read stdin, split at the empty line, parse the colony block with `parser.Parse`, parse the moves into turns
- If stdin starts with `ERROR`, print it and stop

**Think about first**
- The colony block is exactly the original file, so what can you reuse? What do you get for free by having written `Parse(io.Reader)`?
- How do you parse `L12-room` when a room name could itself contain digits or letters? (Which character separates ant from room?)

**Acceptance criteria**
- [ ] Reads the output of every valid audit example
- [ ] An `ERROR:` first line is passed through, with no panic

**You will learn:** reusing a package across two programs.

---

### LEM-23 · Visualizer: layout rooms and tunnels
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| dpavaloa *(proposed)* | dsakkas | LEM-22 | L |

**Goal:** draw the colony on a character grid from the room coordinates.

**Deliverables**
- Scale the coordinates to a terminal-sized grid; draw rooms with their names; draw tunnels between them; mark start and end

**Think about first**
- Coordinates can be negative or large. How do you map them to a grid of 80×24? What if two rooms land on the same cell?
- How do you draw a line between two cells with characters? Which cases are hard (diagonal)? What limitation do you accept and document?

**Acceptance criteria**
- [ ] example00, example04 and example05 are readable
- [ ] Two rooms never overwrite each other silently

**You will learn:** coordinate transforms, drawing with characters.

---

### LEM-24 · Visualizer: animate the turns
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| npanagopoul *(proposed)* | dpavaloa | LEM-23 | M |

**Goal:** show the ants moving, one frame per turn.

**Deliverables**
- Redraw the grid for every turn with ANSI escape codes; ants shown in their rooms; a `-delay` flag; a final line with the number of turns

**Think about first**
- How do you clear and redraw a terminal without flicker?
- When many ants sit in start or end, how do you show that? (A count?)
- Where does `time.Sleep` belong, and why does it not need goroutines? (PRD 11.)

**Acceptance criteria**
- [ ] The ants of example00 visibly move from start to end
- [ ] `-delay 0` shows the final state at once

**You will learn:** terminal output, small user-facing tools.

---

## 10. Wave 4: Finish

### LEM-25 · README
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| TEAM | TEAM | LEM-19 | M |

**Goal:** a README that lets a stranger build, run, understand and test the project.

**Deliverables** (each coder writes their own section; the others review it)
- What it is, how to build and run, input and output examples
- The algorithm in your own words, with one diagram (use example01 to show why the shortest path loses)
- The package map and what each package promises
- How to run the tests and the benchmarks, with measured times
- The bonus items, and the team

**Think about first**
- Can a reader who has never seen the subject follow the algorithm section?
- Which part could *you* not explain in a README? That is the part to revisit.

**Acceptance criteria**
- [ ] A person who did not write it follows it from clone to a working run
- [ ] The algorithm section matches the code

---

### LEM-26 · Audit dry run
| Owner | Reviewer | Depends on | Size |
|---|---|---|---|
| TEAM | TEAM | all | M |

**Goal:** run the audit before the audit.

**Deliverables**
- One team member plays the auditor and asks every question of the audit sheet; another answers and shows it on the screen
- Each coder is asked to explain their own package without looking at the code, and the algorithm on a whiteboard
- Every failed line becomes a new card

**Checklist** (from the audit sheet)

- [ ] Only standard packages used
- [ ] example00 within 6 turns, example01 within 8, example02 within 11, example03 within 6, example04 within 6, example05 within 8
- [ ] The program accepts only `##start` and `##end` as commands
- [ ] Output format correct: one line per turn, `Lx-y` moves
- [ ] badexample00 and badexample01 print the expected error
- [ ] example06 (100 ants) under 1.5 min; example07 (1000 ants) under 2.5 min
- [ ] Ants alone in each room; each tunnel used once per turn
- [ ] All ants end in the `##end` room
- [ ] Repeated runs give the same, correct result
- [ ] Every error case prints a message; the program never crashes
- [ ] Bonus: visualizer; ants visibly move
- [ ] Bonus: specific error messages
- [ ] Bonus: tests exist and check each case; good practices followed
- [ ] Each coder can explain: why the shortest path is not always the answer

---

## 11. Risks and Watch-outs

| Risk | Where it bites | What to do |
|---|---|---|
| Shortest path blocks better routes (example01) | LEM-10 | Draw it by hand first; test the "trap" graph |
| Output depends on Go `map` order | LEM-08, 11, 15 | Iterate over slices or sorted keys; test determinism (LEM-19) |
| `Turns` uses a formula that is wrong when ants < paths | LEM-13 | Exact greedy; test with 1 ant |
| `Assign` and `Turns` drift apart | LEM-14 | The consistency test is mandatory |
| Extra `exit status 1` line in `go run .` | LEM-18 | No `os.Exit(1)`; test the output of bad files |
| Audit page has typos in expected output (`L16 -1`) | LEM-19 | Compare turn count and rules, not text |
| One coder blocks the others | all | LEM-02 on day one; code against contracts with hand-made inputs |
| Checker shares code with the solver | LEM-16 | Independent; reviewer tries to beat it |

---

## 12. Contract Change Log

Any change to a signature in PRD section 8 needs all three coders. Record it here.

| Date | Change | Agreed by |
|---|---|---|
| | *(none yet)* | |

---

*Draft v1.0 · Matches `prd.md` v1.0 · Team: dpavaloa, dsakkas, npanagopoul*