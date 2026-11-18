package output

import (
	"strconv"
	"strings"
	"testing"

	"lem-in/internal/graph"
	"lem-in/internal/parser"
	"lem-in/internal/paths"
	"lem-in/internal/sim"
)

// subjectInput is the file of Example 1 of the subject, comments included.
var subjectInput = []string{
	"3", "##start", "1 23 3", "2 16 7", "#comment", "3 16 3", "4 16 5", "5 9 3",
	"6 1 5", "7 4 8", "##end", "0 9 5", "0-4", "0-6", "1-3", "4-3", "5-2", "3-5",
	"#another comment", "4-2", "2-1", "7-6", "7-2", "7-4", "6-5",
}

// subjectMoves are the turns that the subject prints for Example 1.
var subjectMoves = []string{
	"L1-3 L2-2",
	"L1-4 L2-5 L3-3",
	"L1-0 L2-6 L3-4",
	"L2-0 L3-0",
}

// subjectOutput is the complete text the subject prints for Example 1: the
// input, one empty line, the turns. Written by hand, character by character,
// so that the tests compare against the subject and not against our own code.
func subjectOutput() string {
	text := strings.Join(subjectInput, "\n") + "\n\n"
	return text + strings.Join(subjectMoves, "\n") + "\n"
}

// parseTurns builds turns from text written like the output, for example
// "L1-3 L2-2", so the tests read like a real output.
func parseTurns(lines ...string) []sim.Turn {
	var all []sim.Turn
	for _, line := range lines {
		var turn sim.Turn
		for _, word := range strings.Fields(line) {
			ant, room, _ := strings.Cut(strings.TrimPrefix(word, "L"), "-")
			n, err := strconv.Atoi(ant)
			if err != nil {
				panic("bad test input: " + word)
			}
			turn = append(turn, sim.Move{Ant: n, Room: room})
		}
		all = append(all, turn)
	}
	return all
}

// The card: Example 1 of the subject, character for character.
func TestRenderSubjectExample(t *testing.T) {
	got := Render(subjectInput, parseTurns(subjectMoves...))
	if want := subjectOutput(); got != want {
		t.Fatalf("output differs:\n got %q\nwant %q", got, want)
	}
}

// The whole chain on the subject's own file: parse it, find the paths, move the
// ants, render. The program must print exactly what the subject prints. The
// same file with Windows line endings must give the same output.
func TestRenderWholePipeline(t *testing.T) {
	text := strings.Join(subjectInput, "\n") + "\n"
	inputs := map[string]string{
		"unix line endings":    text,
		"windows line endings": strings.ReplaceAll(text, "\n", "\r\n"),
	}
	for name, input := range inputs {
		t.Run(name, func(t *testing.T) {
			colony, lines, err := parser.Parse(strings.NewReader(input))
			if err != nil {
				t.Fatal(err)
			}
			sets, err := paths.Sets(graph.New(colony), colony.Start, colony.End)
			if err != nil {
				t.Fatal(err)
			}
			best := paths.Best(sets, colony.Ants)

			got := Render(lines, sim.Run(best, colony.Ants))

			if want := subjectOutput(); got != want {
				t.Fatalf("output differs:\n got %q\nwant %q", got, want)
			}
		})
	}
}

// The comments of the input must appear in the output (audit example05 prints
// its "#rooms" comment).
func TestRenderKeepsComments(t *testing.T) {
	got := Render(subjectInput, parseTurns(subjectMoves...))
	for _, comment := range []string{"#comment\n", "#another comment\n"} {
		if !strings.Contains(got, comment) {
			t.Fatalf("the comment %q is missing from the output", comment)
		}
	}
}

// The format rules, checked on the output itself: no space at the end of a
// line, no carriage return, exactly one empty line, and a newline at the very
// end.
func TestRenderFormat(t *testing.T) {
	got := Render(subjectInput, parseTurns(subjectMoves...))

	if !strings.HasSuffix(got, "\n") || strings.HasSuffix(got, "\n\n") {
		t.Fatalf("the output must end with exactly one newline: %q", got)
	}
	if strings.Contains(got, "\r") {
		t.Fatal("the output contains a carriage return")
	}

	emptyLines := 0
	for _, line := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
		if line == "" {
			emptyLines++
		}
		if strings.HasSuffix(line, " ") {
			t.Fatalf("a line ends with a space: %q", line)
		}
	}
	if emptyLines != 1 {
		t.Fatalf("%d empty lines, want exactly 1", emptyLines)
	}
}

// With no turns, the output is the input and the empty line.
func TestRenderNoTurns(t *testing.T) {
	got := Render([]string{"1", "a"}, nil)
	if want := "1\na\n\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// Render prints the turns in the order it gets them. It does not sort.
func TestRenderDoesNotReorder(t *testing.T) {
	got := Render([]string{"x"}, parseTurns("L2-b L1-a"))
	if want := "x\n\nL2-b L1-a\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
