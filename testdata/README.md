# testdata

Input files for the tests of `lem-in`. Every test that needs a colony reads it from here.

Each section below shows the colony as a small drawing. The drawings are schematic: they show
which rooms are linked, not the real coordinates. `start` and `end` mark the `##start` and `##end`
rooms.

To check every file by hand, from the root of the repository:

```bash
go build -o lem-in . && for f in testdata/*.txt; do
  echo "== $f"
  ./lem-in "$f" | awk '/^$/{b=1;next} b{n++} END{print "turns:", n+0}'
  ./lem-in "$f" | grep ERROR
done
```

## Summary

### Audit files (copied exactly from the audit page, do not edit)

| File | Ants | Expected turns | Audit limit |
|---|---|---|---|
| `example00.txt` | 4 | 6 | 6 |
| `example01.txt` | 10 | 8 | 8 |
| `example02.txt` | 20 | 11 | 11 |
| `example03.txt` | 4 | 6 | 6 |
| `example04.txt` | 9 | 6 | 6 |
| `example05.txt` | 9 | 8 | 8 |
| `example06.txt` | 100 | 52 | time only |
| `example07.txt` | 1000 | 502 | time only |
| `badexample00.txt` | 0 | `ERROR: ... invalid number of Ants` | must be rejected |
| `badexample01.txt` | 20 | `ERROR: ... room linked to itself` | must be rejected |

The audit page has a typo in the expected output of example02 (`L16 -1`). That is why the tests
compare the **number of turns** and check the **rules**, and do not compare the text of the output.

### Our own cases

| File | Ants | Expected | Why it exists |
|---|---|---|---|
| `subject_example1.txt` | 3 | 4 turns | Example 1 of the subject, character for character |
| `direct_tunnel.txt` | 3 | 3 turns | start linked straight to end: one tunnel can carry one ant per turn |
| `one_ant.txt` | 1 | 3 turns | the smallest real colony: one ant, one path |
| `fewer_ants_than_paths.txt` | 2 | 2 turns | more paths than ants: the extra path must stay unused |
| `comments_and_unknown_command.txt` | 2 | 3 turns | comments and an unknown `##command` are ignored, and comments are echoed |
| `rooms_before_start.txt` | 2 | 4 turns | rooms declared before `##start` are normal rooms |
| `word_names.txt` | 2 | 3 turns | room names are words, not numbers |
| `shortest_path_trap.txt` | 2 | 3 turns | the shortest path is not part of the best answer |
| `windows_line_endings.txt` | 1 | 1 turn | `\r\n` line endings must not appear in the output |
| `bad_no_path.txt` | 1 | `ERROR: ... no path between start and end` | rooms exist, but nothing joins start to end |
| `bad_no_start.txt` | 1 | `ERROR: ... no start room found` | `##start` is missing |
| `bad_duplicate_room.txt` | 1 | `ERROR: ... duplicated room` | two rooms with the same name |
| `bad_unknown_link.txt` | 1 | `ERROR: ... link to unknown room` | a tunnel names a room that does not exist |
| `bad_room_name_L.txt` | 1 | `ERROR: ... invalid room name` | a room name must not start with `L` |
| `bad_coordinates.txt` | 1 | `ERROR: ... invalid coordinates` | a coordinate must be an integer |

(`...` stands for `invalid data format,`.)

## Audit colonies

### example00: one path of three moves

```
[0 start]──[2]──[3]──[1 end]
```

One path, length 3. Four ants go one behind the other: 3 + 4 - 1 = 6 turns.

### example01: three lanes with cross links

```
          ┌──[t]──[E]──[a]──[m]──┐
[start]───┼──[0]──[o]──[n]──[e]──┼───[end]
          └──[h]──[A]──[c]──[k]──┘
```

Extra links: `h-n` joins the bottom lane to the middle lane, and `n-m` joins the middle lane to the
top lane.

The shortest path is `start-h-n-m-end` (4 moves), but it cuts across all three lanes and blocks
them. The best answer is the three lanes of 5 moves each: 10 ants take 8 turns. This is the file
that proves the shortest path is not part of the answer.

### example02: a direct tunnel and a long way round

```
[0 start]────────────────[3 end]
    │                       │
   [1]──────────────────[2]
```

Two paths: `0-3` (1 move) and `0-1-2-3` (3 moves). With 20 ants the best split is 11 and 9 ants,
which gives 11 turns. This is the file with the typo in the audit's expected output.

### example03: everything goes through one room

```
          ┌──[1]──┐
[0 start]─┼──[2]──┼──[4]──[5 end]
          └──[3]──┘
```

Room `4` is declared before `##start` and is the only room linked to the end. So there is one
useful path and four ants wait for it: 3 + 4 - 1 = 6 turns.

### example04: two paths out of six rooms

```
                 [gilfoyle]
               ╱      │      ╲
[richard]─────╱─────[erlich]──╲────[peter]
     │                          ╲   ╱
     └─────[dinish]────────[jimYoung]
```

Links: `richard-dinish`, `dinish-jimYoung`, `richard-gilfoyle`, `gilfoyle-peter`,
`gilfoyle-erlich`, `richard-erlich`, `erlich-jimYoung`, `jimYoung-peter`.

Best answer: `richard-gilfoyle-peter` (2 moves) and `richard-dinish-jimYoung-peter` (3 moves).
9 ants take 6 turns.

### example05: a large colony with a comment

```
        ┌──[C0]─[C1]─[C2]─[C3]─[I4]─[I5]──────────────────┐
        ├──[A0]─[A1]─[A2]─────────────────────────────────┤
[start]─┤                                                 ├─[end]
        ├──[B0]─[B1]──[E2]─[D2]─[D3]──────────────────────┤
        └──[G0]─[G1]─[G2]─[G3]─[G4]───[D3]                │
                                       [F2]─[F3]─[F4]─────┘
```

This drawing only shows the main lanes. The exact links are in the file. Several cross links join
the lanes (`A0-D1`, `A1-B1`, `A2-C3`, `D1-D2`, `D2-F3`, `H3-H4`, ...).

The file starts with a `#rooms` comment. The comment must appear in the output. Best answer:
three paths, 9 ants, 8 turns.

### example06 and example07: the same colony with many ants

The colony of example04 with 100 and with 1000 ants. They are small: the point is the **time**
(the audit limits are 1.5 and 2.5 minutes, our targets are 0.1 s and 0.5 s). Expected turns: 52
and 502.

### badexample00: zero ants

```
0
4 5 4
...
```

The same colony as example03, but the ant count is `0`. It must be rejected with
`invalid number of Ants`.

### badexample01: a tunnel to itself

```
...
2-3
3-3     <- a room linked to itself
...
7-8
8-7     <- the same tunnel twice
```

The file has several mistakes: `3-3` (self link), `7-8` and `8-7` (duplicate tunnel) and no
tunnel that reaches the end room. The parser reports the first one it meets, which is
`room linked to itself`. Any of them is enough to reject the file.

## Our own colonies

### subject_example1: Example 1 of the subject

```
          ┌──[3]──[4]──────┐
[1 start]─┤                ├──[0 end]
          └──[2]──[5]──[6]─┘
```

Extra links: `4-2`, `3-5`, `7-6`, `7-2`, `7-4`. Two paths are used: `1-3-4-0` and `1-2-5-6-0`.
3 ants take 4 turns. The output must match the subject exactly.

### direct_tunnel: start and end are neighbours

```
[s start]────[e end]
```

One tunnel, three ants. The tunnel can be used once per turn: 3 turns.

### one_ant: a single ant

```
[s start]──[a]──[b]──[e end]
```

One ant walks three moves: 3 turns.

### fewer_ants_than_paths: more paths than ants

```
          ┌──[a]──┐
[s start]─┼──[b]──┼──[e end]
          └──[c]──┘
```

Three paths of two moves, two ants. Each ant takes its own path: 2 turns. The third path must not
be used.

### comments_and_unknown_command: lines that must be ignored

```
[s start]──[a]──[e end]
```

The file has comments in several places and two unknown commands (`##speed 9`,
`##another_unknown_command`). The colony is a simple chain. 2 ants take 3 turns, and the comments
must be echoed in the output.

### rooms_before_start: order of declaration

```
[s start]──[a]──[b]──[e end]
```

Rooms `a` and `b` are declared **before** `##start`. They are ordinary rooms. 2 ants take 4 turns.

### word_names: names that are not numbers

```
[kitchen start]──[hall]──[garden end]
```

Room names are words. 2 ants take 3 turns.

### shortest_path_trap: the shortest path blocks the answer

```
 s ───── a ───── d ───── e
 │       │               │
 │       b ──────────────┘
 │       │
 └────── c
```

Links: `s-a`, `a-b`, `b-e`, `s-c`, `c-b`, `a-d`, `d-e`.

The shortest path `s-a-b-e` uses `a` and `b`. If it is chosen first, the other ant has no free
route and the two ants need 4 turns. The best answer is `s-a-d-e` and `s-c-b-e`: two separate
paths, 3 turns. A solver that only repeats "find the shortest path, then block its rooms" fails
here. The solver has to be able to undo part of a path it found earlier.

### windows_line_endings: `\r\n`

```
[s start]──[e end]
```

The file is written with `\r\n`. The output must not contain `\r`, and the result must be the
same as for the same file with `\n`.

### bad_no_path

```
[s start]──[a]        [e end]
```

`e` exists, but no tunnel reaches it.

### bad_no_start

```
##end → [e]──[a]
```

There is a `##end` and no `##start`.

### bad_duplicate_room

```
[s start]  [s]──[e end]
```

Two rooms are both called `s`.

### bad_unknown_link

```
[s start]──[e end]      s-x → ???
```

The tunnel `s-x` names a room, `x`, that was never declared.

### bad_room_name_L

```
[Ls start]──[e end]
```

A room name must not start with `L`, because `L1-3` is how a move is written.

### bad_coordinates

```
[s start 0 zero]──[e end]
```

The second coordinate is `zero`, not an integer.