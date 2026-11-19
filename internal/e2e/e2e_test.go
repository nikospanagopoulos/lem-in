// Package e2e holds end-to-end tests: they build the real lem-in program, run
// it on the files of testdata/ and judge only what it prints, exactly as an
// auditor does. There is no production code in this package.
package e2e

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"lem-in/internal/graph"
	"lem-in/internal/parser"
	"lem-in/internal/sim"
)

// programPath is where TestMain puts the built program. A package-level
// variable is fine here: it is set once, before any test runs, and only read
// afterwards.
var programPath string

// TestMain runs once for the whole package, before and after all the tests.
// It builds the program a single time instead of once per test, which keeps
// the tests fast. The program is built into a temporary folder that is removed
// at the end, so nothing is left in the repository.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "lem-in-e2e")
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot create temporary folder:", err)
		os.Exit(1)
	}
	programPath = filepath.Join(dir, "lem-in")

	// Tests run with the package folder as the working directory, so the root
	// of the module, where main.go is, is two folders up.
	build := exec.Command("go", "build", "-o", programPath, ".")
	build.Dir = filepath.Join("..", "..")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "go build failed: %v\n%s", err, out)
		os.RemoveAll(dir)
		os.Exit(1)
	}

	code := m.Run()
	// os.Exit does not run deferred calls, so the folder is removed by hand.
	os.RemoveAll(dir)
	os.Exit(code)
}

// trackOnce makes sure the sources are opened only once per test run.
var trackOnce sync.Once

// trackSources opens every .go file and every file of testdata/ under root.
//
// Why: Go caches the result of a test that has not changed and skips running
// it again. It only notices the files that the test opens while it runs, and
// this test builds the program in another process, so a change in the code of
// lem-in would go unnoticed and an old "ok" would be shown. Opening the
// sources here makes Go include them in that decision. It must happen while
// the tests run, not in TestMain before m.Run: Go starts recording opened
// files only inside m.Run.
func trackSources(root string) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" && filepath.Base(filepath.Dir(path)) != "testdata" {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		return file.Close()
	})
}

// runProgram runs lem-in with the given arguments and returns what it printed
// on the standard output. The program never exits with an error status (see
// PRD section 9), so any failure to run it at all is a test failure.
func runProgram(t *testing.T, args ...string) string {
	t.Helper()
	trackOnce.Do(func() {
		if err := trackSources(filepath.Join("..", "..")); err != nil {
			t.Fatalf("cannot read the sources: %v", err)
		}
	})
	cmd := exec.Command(programPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("lem-in %v failed to run: %v\nstderr: %s", args, err, stderr.String())
	}
	if stderr.Len() > 0 {
		t.Fatalf("lem-in %v wrote to stderr: %s", args, stderr.String())
	}
	return stdout.String()
}

// goodCases lists every valid colony and the fewest turns it can take (PRD
// section 6.5 and testdata/README.md). If the program is slower than this, the
// test fails; if it were faster, the number would be wrong.
var goodCases = map[string]int{
	"example00.txt":                    6,
	"example01.txt":                    8,
	"example02.txt":                    11,
	"example03.txt":                    6,
	"example04.txt":                    6,
	"example05.txt":                    8,
	"example06.txt":                    52,
	"example07.txt":                    502,
	"subject_example1.txt":             4,
	"direct_tunnel.txt":                3,
	"one_ant.txt":                      3,
	"fewer_ants_than_paths.txt":        2,
	"comments_and_unknown_command.txt": 3,
	"rooms_before_start.txt":           4,
	"word_names.txt":                   3,
	"shortest_path_trap.txt":           3,
	"windows_line_endings.txt":         1,
}

// badCases lists every file that must be rejected and the exact line expected.
var badCases = map[string]string{
	"badexample00.txt":       "ERROR: invalid data format, invalid number of Ants",
	"badexample01.txt":       "ERROR: invalid data format, room linked to itself",
	"bad_no_path.txt":        "ERROR: invalid data format, no path between start and end",
	"bad_no_start.txt":       "ERROR: invalid data format, no start room found",
	"bad_duplicate_room.txt": "ERROR: invalid data format, duplicated room",
	"bad_unknown_link.txt":   "ERROR: invalid data format, link to unknown room",
	"bad_room_name_L.txt":    "ERROR: invalid data format, invalid room name",
	"bad_coordinates.txt":    "ERROR: invalid data format, invalid coordinates",
}

func testdataPath(name string) string {
	return filepath.Join("..", "..", "testdata", name)
}

// splitOutput cuts the printed text at its one empty line. The part before it
// is the echo of the input, the part after it holds the turns. Each part is
// returned as a list of lines.
func splitOutput(t *testing.T, text string) (echo, turns []string) {
	t.Helper()
	if !strings.HasSuffix(text, "\n") {
		t.Fatal("the output does not end with a newline")
	}
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	blank := -1
	for i, line := range lines {
		if line == "" {
			if blank >= 0 {
				t.Fatalf("a second empty line at line %d", i+1)
			}
			blank = i
		}
	}
	if blank < 0 {
		t.Fatal("the output has no empty line between the input and the turns")
	}
	return lines[:blank], lines[blank+1:]
}

// parseMoves turns the printed turn lines back into sim.Turn values, the same
// type that sim.Check understands. This is why the test reads the TEXT: it
// proves what the user sees, including the formatting, is a valid solution.
func parseMoves(t *testing.T, lines []string) []sim.Turn {
	t.Helper()
	turns := make([]sim.Turn, 0, len(lines))
	for n, line := range lines {
		if line == "" || strings.HasSuffix(line, " ") || strings.Contains(line, "  ") {
			t.Fatalf("turn %d is badly formatted: %q", n+1, line)
		}
		var turn sim.Turn
		for _, word := range strings.Split(line, " ") {
			body, ok := strings.CutPrefix(word, "L")
			if !ok {
				t.Fatalf("turn %d: %q does not start with L", n+1, word)
			}
			// A room name never contains '-', so the first '-' separates the
			// ant number from the room.
			ant, room, ok := strings.Cut(body, "-")
			number, err := strconv.Atoi(ant)
			if !ok || err != nil || room == "" {
				t.Fatalf("turn %d: %q is not of the form L<ant>-<room>", n+1, word)
			}
			turn = append(turn, sim.Move{Ant: number, Room: room})
		}
		turns = append(turns, turn)
	}
	return turns
}

// colonyOf reads a test file with the real parser, to give sim.Check the
// colony the moves must obey.
func colonyOf(t *testing.T, path string) (graph.Colony, []string) {
	t.Helper()
	colony, lines, err := parser.ParseFile(path)
	if err != nil {
		t.Fatalf("the parser rejects %s: %v", path, err)
	}
	return colony, lines
}

// The three assertions of the card, for every valid colony:
//
//	A. the number of turns equals the best known number  (catches a SLOW answer)
//	B. sim.Check accepts the moves                       (catches an INVALID answer)
//	C. two runs print exactly the same bytes             (catches map-order bugs)
func TestGoodColonies(t *testing.T) {
	for name, wantTurns := range goodCases {
		t.Run(name, func(t *testing.T) {
			path := testdataPath(name)
			first := runProgram(t, path)

			echo, turnLines := splitOutput(t, first)
			colony, lines := colonyOf(t, path)

			// The first part must be the input, line for line.
			if strings.Join(echo, "\n") != strings.Join(lines, "\n") {
				t.Fatalf("the echo of the input differs from the input:\n%s", first)
			}

			turns := parseMoves(t, turnLines)

			// A: not too slow.
			if len(turns) != wantTurns {
				t.Errorf("%d turns, want %d", len(turns), wantTurns)
			}
			// B: obeys every rule of the subject.
			if err := sim.Check(colony, turns); err != nil {
				t.Errorf("the moves break a rule: %v", err)
			}
			// C: same bytes every time.
			if second := runProgram(t, path); second != first {
				t.Error("two runs of the same file printed different text")
			}
		})
	}
}

// Every bad file prints exactly one line and nothing else.
func TestBadColonies(t *testing.T) {
	for name, want := range badCases {
		t.Run(name, func(t *testing.T) {
			got := runProgram(t, testdataPath(name))
			if got != want+"\n" {
				t.Fatalf("got %q, want %q", got, want+"\n")
			}
		})
	}
}

// No argument, a missing file and a folder as argument: one ERROR line, no
// panic, nothing else.
func TestBadArguments(t *testing.T) {
	const want = "ERROR: invalid data format, cannot read file\n"
	cases := map[string][]string{
		"no argument":   nil,
		"two arguments": {testdataPath("one_ant.txt"), testdataPath("one_ant.txt")},
		"missing file":  {testdataPath("does_not_exist.txt")},
		"a folder":      {testdataPath("")},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			if got := runProgram(t, args...); got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}

// A new file in testdata/ must be added to the tables above, or this test
// fails. That way a file cannot sit in testdata/ without being checked.
func TestEveryFileIsCovered(t *testing.T) {
	files, err := filepath.Glob(testdataPath("*.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != len(goodCases)+len(badCases) {
		t.Errorf("testdata has %d .txt files, the tables list %d",
			len(files), len(goodCases)+len(badCases))
	}
	for _, file := range files {
		name := filepath.Base(file)
		_, good := goodCases[name]
		_, bad := badCases[name]
		if !good && !bad {
			t.Errorf("%s is in testdata/ but not in goodCases or badCases", name)
		}
	}
}
