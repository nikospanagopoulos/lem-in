package main

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// One frame before the first turn, one after each turn, and a last line.
func TestRunShowsOneFramePerTurn(t *testing.T) {
	var out strings.Builder
	err := run([]string{"-delay", "0", "-color=false"}, strings.NewReader(subjectOutput), &out)
	if err != nil {
		t.Fatal(err)
	}
	got := out.String()

	// Example 1 of the subject has 4 turns: frames 0, 1, 2, 3 and 4.
	if n := strings.Count(got, clearScreen); n != 5 {
		t.Errorf("%d frames, want 5", n)
	}
	for turn := 0; turn <= 4; turn++ {
		title := "Turn " + string(rune('0'+turn)) + "/4\n"
		if !strings.Contains(got, title) {
			t.Errorf("missing %q", title)
		}
	}
	if !strings.HasSuffix(got, "end [0]: 3 ants\nants in rooms: none\nDone: 3 ants reached the end in 4 turns.\n") {
		t.Errorf("the output does not end as expected:\n%q", got[len(got)-120:])
	}
	if strings.Contains(got, "\x1b[1;") || strings.Contains(got, "\x1b[3") {
		t.Error("colours are on although -color=false was given")
	}
}

// Colours are on by default.
func TestRunUsesColoursByDefault(t *testing.T) {
	var out strings.Builder
	if err := run([]string{"-delay", "0"}, strings.NewReader(subjectOutput), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), ansiStart) || !strings.Contains(out.String(), ansiAnts) {
		t.Error("no colour codes in the output")
	}
}

// The delay is a pause before each turn, so one turn takes at least that long.
func TestRunWaitsBetweenFrames(t *testing.T) {
	const colony = "1\n##start\ns 0 0\n##end\ne 1 0\ns-e\n\nL1-e\n"

	started := time.Now()
	var out strings.Builder
	if err := run([]string{"-delay", "40ms"}, strings.NewReader(colony), &out); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(started); elapsed < 40*time.Millisecond {
		t.Errorf("finished after %v, the delay of 40ms was not respected", elapsed)
	}
}

// A colony with no moves still shows its one frame.
func TestRunWithNoTurns(t *testing.T) {
	const colony = "1\n##start\ns 0 0\n##end\ne 1 0\ns-e\n\n"
	var out strings.Builder
	if err := run([]string{"-delay", "0", "-color=false"}, strings.NewReader(colony), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Turn 0/0\n") || !strings.HasSuffix(out.String(), "in 0 turns.\n") {
		t.Errorf("unexpected output: %q", out.String())
	}
}

// What lem-in prints when it rejects a file is passed on as it is, and nothing
// is drawn.
func TestRunSolverError(t *testing.T) {
	var out strings.Builder
	err := run(nil, strings.NewReader("ERROR: invalid data format, no start room found\n"), &out)

	var solver *solverError
	if !errors.As(err, &solver) {
		t.Fatalf("error = %v, want a *solverError", err)
	}
	if out.Len() != 0 {
		t.Errorf("something was drawn: %q", out.String())
	}
}

func TestRunBadArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"unknown flag", []string{"-speed", "3"}},
		{"delay that is not a duration", []string{"-delay", "fast"}},
		{"negative delay", []string{"-delay", "-1s"}},
		{"width that is not a number", []string{"-width", "wide"}},
		{"a file name", []string{"colony.txt"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out strings.Builder
			err := run(tt.args, strings.NewReader(subjectOutput), &out)
			if err == nil {
				t.Fatal("no error")
			}
			if out.Len() != 0 {
				t.Errorf("something was drawn: %q", out.String())
			}
		})
	}
}

func TestRunAreaTooSmall(t *testing.T) {
	for _, args := range [][]string{{"-width", "2"}, {"-height", "0"}} {
		var out strings.Builder
		err := run(args, strings.NewReader(subjectOutput), &out)
		if !errors.Is(err, errTooSmall) {
			t.Errorf("%v: error = %v, want errTooSmall", args, err)
		}
	}
}

// A move of an ant that does not exist stops the program with an error that
// says which turn it was.
func TestRunUnknownAnt(t *testing.T) {
	const colony = "1\n##start\ns 0 0\n##end\ne 1 0\ns-e\n\nL5-e\n"
	var out strings.Builder
	err := run([]string{"-delay", "0"}, strings.NewReader(colony), &out)
	if !errors.Is(err, errUnknownAnt) || !strings.Contains(err.Error(), "turn 1") {
		t.Fatalf("error = %v, want errUnknownAnt for turn 1", err)
	}
}
