package main

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"lem-in/internal/graph"
	"lem-in/internal/output"
	"lem-in/internal/parser"
	"lem-in/internal/paths"
	"lem-in/internal/sim"
)

// subjectOutput is what lem-in prints for Example 1 of the subject, written by
// hand from the subject (not built by our own code).
const subjectOutput = `3
##start
1 23 3
2 16 7
#comment
3 16 3
4 16 5
5 9 3
6 1 5
7 4 8
##end
0 9 5
0-4
0-6
1-3
4-3
5-2
3-5
#another comment
4-2
2-1
7-6
7-2
7-4
6-5

L1-3 L2-2
L1-4 L2-5 L3-3
L1-0 L2-6 L3-4
L2-0 L3-0
`

func TestReadInputSubjectExample(t *testing.T) {
	got, err := readInput(strings.NewReader(subjectOutput))
	if err != nil {
		t.Fatal(err)
	}

	if got.Colony.Ants != 3 || got.Colony.Start != "1" || got.Colony.End != "0" {
		t.Errorf("colony read wrongly: %+v", got.Colony)
	}
	if len(got.Colony.Rooms) != 8 || len(got.Colony.Links) != 12 {
		t.Errorf("%d rooms and %d tunnels, want 8 and 12",
			len(got.Colony.Rooms), len(got.Colony.Links))
	}

	want := []sim.Turn{
		{{Ant: 1, Room: "3"}, {Ant: 2, Room: "2"}},
		{{Ant: 1, Room: "4"}, {Ant: 2, Room: "5"}, {Ant: 3, Room: "3"}},
		{{Ant: 1, Room: "0"}, {Ant: 2, Room: "6"}, {Ant: 3, Room: "4"}},
		{{Ant: 2, Room: "0"}, {Ant: 3, Room: "0"}},
	}
	if !reflect.DeepEqual(got.Turns, want) {
		t.Errorf("turns:\n got %v\nwant %v", got.Turns, want)
	}
}

// The visualizer must read exactly what the solver writes. This runs the real
// solver pipeline, renders it, reads it back and compares with the turns the
// solver started from.
func TestReadInputReadsWhatTheSolverWrites(t *testing.T) {
	// The input file is the first block of the subject's output.
	inputText, _, _ := strings.Cut(subjectOutput, "\n\n")
	colony, lines, err := parser.Parse(strings.NewReader(inputText))
	if err != nil {
		t.Fatal(err)
	}
	sets, err := paths.Sets(graph.New(colony), colony.Start, colony.End)
	if err != nil {
		t.Fatal(err)
	}
	turns := sim.Run(paths.Best(sets, colony.Ants), colony.Ants)

	got, err := readInput(strings.NewReader(output.Render(lines, turns)))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Turns, turns) {
		t.Errorf("turns read back differ:\n got %v\nwant %v", got.Turns, turns)
	}
}

// An ERROR line from lem-in is passed on as it is.
func TestReadInputSolverError(t *testing.T) {
	const line = "ERROR: invalid data format, no start room found"
	_, err := readInput(strings.NewReader(line + "\n"))

	var solver *solverError
	if !errors.As(err, &solver) {
		t.Fatalf("error = %v, want a *solverError", err)
	}
	if solver.Line != line {
		t.Errorf("line = %q, want %q", solver.Line, line)
	}
}

func TestReadInputEdgeCases(t *testing.T) {
	const colony = "1\n##start\ns 0 0\n##end\ne 1 0\ns-e\n"

	tests := []struct {
		name    string
		text    string
		wantErr error // nil means the input must be accepted
		turns   int   // number of turns expected when accepted
	}{
		{"one turn", colony + "\nL1-e\n", nil, 1},
		{"no turns", colony + "\n", nil, 0},
		{"extra empty lines at the end", colony + "\nL1-e\n\n\n", nil, 1},
		{"windows line endings", strings.ReplaceAll(colony+"\nL1-e\n", "\n", "\r\n"), nil, 1},
		{"empty input", "", errEmptyInput, 0},
		{"starts with an empty line", "\n\nL1-e\n", parser.ErrAnts, 0},
		{"only one line", "1\n", errNoSeparator, 0},
		{"no empty line", colony, errNoSeparator, 0},
		{"colony is not valid", "0\n##start\ns 0 0\n##end\ne 1 0\ns-e\n\nL1-e\n", parser.ErrAnts, 0},
		{"move without L", colony + "\n1-e\n", errBadMove, 0},
		{"move without a room", colony + "\nL1-\n", errBadMove, 0},
		{"move without a dash", colony + "\nL1e\n", errBadMove, 0},
		{"ant is not a number", colony + "\nLx-e\n", errBadMove, 0},
		{"ant number zero", colony + "\nL0-e\n", errBadMove, 0},
		{"empty turn in the middle", colony + "\nL1-e\n\nL1-e\n", errBadMove, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readInput(strings.NewReader(tt.text))
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got.Turns) != tt.turns {
				t.Errorf("%d turns, want %d", len(got.Turns), tt.turns)
			}
		})
	}
}

// A room whose name contains letters and digits, and an ant number with two
// digits, are read correctly.
func TestParseMove(t *testing.T) {
	got, err := parseMove("L12-jimYoung")
	if err != nil {
		t.Fatal(err)
	}
	if want := (sim.Move{Ant: 12, Room: "jimYoung"}); got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
