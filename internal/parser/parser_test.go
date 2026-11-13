package parser

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"lem-in/internal/graph"
)

// file joins lines into the text of an input file.
// The "..." makes it variadic: file("a", "b") works, with no slice needed.
// Writing test inputs as separate lines is much easier to read than one long
// string full of \n.
func file(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}

// valid is a minimal correct colony without its first line (the ant count).
// Tests that need only one thing to be wrong build their input from it.
var valid = []string{"##start", "a 0 0", "##end", "b 1 1", "a-b"}

// withAnts is a correct colony whose first line is ants.
func withAnts(ants string) string {
	// We build a NEW slice that starts with the ant count and continues with
	// valid. We never do append(valid, ...) on valid itself: append may write
	// into the shared array and silently change valid for every other test.
	return file(append([]string{ants}, valid...)...)
}

// subjectLines is the colony of Example 1 of the subject, comments included.
var subjectLines = []string{
	"3",
	"##start",
	"1 23 3",
	"2 16 7",
	"#comment",
	"3 16 3",
	"4 16 5",
	"5 9 3",
	"6 1 5",
	"7 4 8",
	"##end",
	"0 9 5",
	"0-4",
	"0-6",
	"1-3",
	"4-3",
	"5-2",
	"3-5",
	"#another comment",
	"4-2",
	"2-1",
	"7-6",
	"7-2",
	"7-4",
	"6-5",
}

// The most important test: the exact example of the subject must give exactly
// the colony we expect, and the raw lines must come back untouched.
func TestParseSubjectExample(t *testing.T) {
	// strings.NewReader turns a string into an io.Reader, so Parse can be
	// tested without creating a file on disk.
	colony, lines, err := Parse(strings.NewReader(file(subjectLines...)))
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}

	want := graph.Colony{
		Ants:  3,
		Start: "1",
		End:   "0",
		Rooms: []graph.Room{
			{Name: "1", X: 23, Y: 3}, {Name: "2", X: 16, Y: 7},
			{Name: "3", X: 16, Y: 3}, {Name: "4", X: 16, Y: 5},
			{Name: "5", X: 9, Y: 3}, {Name: "6", X: 1, Y: 5},
			{Name: "7", X: 4, Y: 8}, {Name: "0", X: 9, Y: 5},
		},
		Links: [][2]string{
			{"0", "4"}, {"0", "6"}, {"1", "3"}, {"4", "3"}, {"5", "2"}, {"3", "5"},
			{"4", "2"}, {"2", "1"}, {"7", "6"}, {"7", "2"}, {"7", "4"}, {"6", "5"},
		},
	}
	// A struct that contains slices cannot be compared with ==, so we use
	// reflect.DeepEqual, which compares the content element by element.
	// Note that the order of the rooms and tunnels matters: it must be the
	// order of the file.
	if !reflect.DeepEqual(colony, want) {
		t.Fatalf("colony = %+v\nwant     %+v", colony, want)
	}
	if !slices.Equal(lines, subjectLines) {
		t.Fatalf("raw lines differ from the input:\n got %q\nwant %q", lines, subjectLines)
	}
}

// Inputs that look unusual but are valid. Each row says what Parse should find.
// The loop at the bottom is the same table-driven pattern as in graph_test.go:
// a struct describes one case, and t.Run runs it under its own name.
func TestParseAcceptsValidInputs(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantAnts  int
		wantStart string
		wantEnd   string
		wantRooms int
		wantLinks int
	}{
		{
			name: "rooms declared before ##start (audit example03)",
			input: file("4", "4 5 4", "##start", "0 1 4", "1 3 6", "##end", "5 6 4",
				"2 3 4", "3 3 1", "0-1", "2-4", "1-4", "0-2", "4-5", "3-0", "4-3"),
			wantAnts: 4, wantStart: "0", wantEnd: "5", wantRooms: 6, wantLinks: 7,
		},
		{
			name: "comments and unknown commands are ignored",
			input: file("2", "#hi", "##unknown", "##start", "#note", "a 0 0", "# between",
				"##end", "b 1 1", "##another", "a-b"),
			wantAnts: 2, wantStart: "a", wantEnd: "b", wantRooms: 2, wantLinks: 1,
		},
		{
			name:     "##end before ##start",
			input:    file("1", "##end", "b 1 1", "##start", "a 0 0", "a-b"),
			wantAnts: 1, wantStart: "a", wantEnd: "b", wantRooms: 2, wantLinks: 1,
		},
		{
			name:     "tunnel before the rooms: order does not matter",
			input:    file("1", "a-b", "##start", "a 0 0", "##end", "b 1 1"),
			wantAnts: 1, wantStart: "a", wantEnd: "b", wantRooms: 2, wantLinks: 1,
		},
		{
			name:     "negative coordinates",
			input:    file("1", "##start", "a -1 -2", "##end", "b 1 1", "a-b"),
			wantAnts: 1, wantStart: "a", wantEnd: "b", wantRooms: 2, wantLinks: 1,
		},
		{
			name:     "ant count with a leading zero",
			input:    withAnts("03"),
			wantAnts: 3, wantStart: "a", wantEnd: "b", wantRooms: 2, wantLinks: 1,
		},
		{
			name:     "room names that are words",
			input:    file("1", "##start", "richard 0 6", "##end", "peter 14 6", "richard-peter"),
			wantAnts: 1, wantStart: "richard", wantEnd: "peter", wantRooms: 2, wantLinks: 1,
		},
		{
			name:     "no newline at the end of the file",
			input:    strings.TrimSuffix(withAnts("5"), "\n"),
			wantAnts: 5, wantStart: "a", wantEnd: "b", wantRooms: 2, wantLinks: 1,
		},
		{
			name:     "Windows line endings",
			input:    strings.ReplaceAll(withAnts("5"), "\n", "\r\n"),
			wantAnts: 5, wantStart: "a", wantEnd: "b", wantRooms: 2, wantLinks: 1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			colony, lines, err := Parse(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("Parse returned an error: %v", err)
			}
			if colony.Ants != tc.wantAnts || colony.Start != tc.wantStart || colony.End != tc.wantEnd ||
				len(colony.Rooms) != tc.wantRooms || len(colony.Links) != tc.wantLinks {
				t.Fatalf("colony = %+v", colony)
			}
			// Whatever the line endings were, no raw line may keep the
			// "\r": we print these lines back, and a stray "\r" would
			// corrupt the output.
			for _, line := range lines {
				if strings.Contains(line, "\r") {
					t.Fatalf("a raw line still has a carriage return: %q", line)
				}
			}
		})
	}
}

// The raw lines are what the program prints before the moves, so they must be
// the input as it was: comments included, trailing blank lines removed.
func TestParseKeepsRawLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"plain file", file(subjectLines...), subjectLines},
		{"blank lines at the end are dropped", file(subjectLines...) + "\n\n  \n", subjectLines},
		{"no newline at the end", strings.TrimSuffix(file(subjectLines...), "\n"), subjectLines},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, lines, err := Parse(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("Parse returned an error: %v", err)
			}
			if !slices.Equal(lines, tc.want) {
				t.Fatalf("raw lines:\n got %q\nwant %q", lines, tc.want)
			}
		})
	}
}

// audit badexample01: a room linked to itself, a tunnel written twice, and
// more.
var auditBadExample01 = file(
	"20", "##start", "0 2 0", "1 7 0", "2 13 0", "3 18 0", "##end", "4 23 0",
	"5 7 6", "6 10 6", "7 13 6", "8 16 6", "9 7 2", "10 7 4", "11 13 2",
	"12 15 2", "13 17 2", "14 19 2", "15 21 2", "16 9 4",
	"0-1", "0-5", "0-9", "0-10", "1-2", "2-3", "3-3", "5-6", "6-7", "7-8",
	"8-7", "9-2", "10-16", "1-11", "11-12", "12-13", "13-14", "14-15", "15-1", "16-7",
)

// audit badexample00: zero ants.
var auditBadExample00 = file(
	"0", "4 5 4", "##start", "0 1 4", "1 3 6", "##end", "5 6 4", "2 3 4", "3 3 1",
	"0-1", "2-4", "1-4", "0-2", "4-5", "3-0", "4-3",
)

// TestParseErrors has one case for every row of the error table of the PRD
// (section 9), plus the two bad examples of the audit.
//
// errors.Is checks WHICH error came back, not its text. That is the point of
// the sentinel errors in errors.go: if we reword a message, no test breaks.
// The cases are grouped by the part of the file that is wrong.
func TestParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  error
	}{
		// Number of ants.
		{"empty file", "", ErrAnts},
		{"only blank lines", "\n\n\n", ErrAnts},
		{"ants: a word", withAnts("three"), ErrAnts},
		{"ants: zero", withAnts("0"), ErrAnts},
		{"ants: negative", withAnts("-1"), ErrAnts},
		{"ants: with a plus sign", withAnts("+3"), ErrAnts},
		{"ants: leading space", withAnts(" 3"), ErrAnts},
		{"ants: trailing space", withAnts("3 "), ErrAnts},
		{"ants: decimal", withAnts("3.5"), ErrAnts},
		{"ants: too large for an int", withAnts("99999999999999999999"), ErrAnts},
		{"ants: first line is a command", withAnts("##start"), ErrAnts},
		{"ants: blank first line", "\n" + withAnts("3"), ErrAnts},
		{"audit badexample00 (zero ants)", auditBadExample00, ErrAnts},

		// Start and end.
		{"no ##start", file("1", "##end", "a 0 0", "b 1 1", "a-b"), ErrNoStart},
		{"##start followed by a tunnel", file("1", "##start", "a-b", "a 0 0", "##end", "b 1 1"), ErrNoStart},
		{"##start at the end of the file", file("1", "##end", "b 1 1", "##start"), ErrNoStart},
		{"##start followed by ##end", file("1", "##start", "##end", "a 0 0", "b 1 1"), ErrNoStart},
		{"no ##end", file("1", "##start", "a 0 0", "b 1 1", "a-b"), ErrNoEnd},
		{"##end followed by a tunnel", file("1", "##start", "a 0 0", "##end", "a-b", "b 1 1"), ErrNoEnd},
		{"##end at the end of the file", file("1", "##start", "a 0 0", "##end"), ErrNoEnd},
		{"##end followed by ##start", file("1", "##end", "##start", "a 0 0", "b 1 1"), ErrNoEnd},
		{"two ##start", file("1", "##start", "a 0 0", "##start", "b 1 1", "##end", "c 2 2"), ErrDuplicateStartEnd},
		{"two ##end", file("1", "##start", "a 0 0", "##end", "b 1 1", "##end", "c 2 2"), ErrDuplicateStartEnd},

		// Rooms.
		{"duplicated room", file("1", "##start", "a 0 0", "##end", "b 1 1", "a 5 5"), ErrDuplicateRoom},
		{"start and end are the same room", file("1", "##start", "a 0 0", "##end", "a 1 1"), ErrDuplicateRoom},
		{"room name starts with L", file("1", "##start", "a 0 0", "##end", "b 1 1", "Lroom 2 2"), ErrRoomName},
		{"room name has a dash", file("1", "##start", "a 0 0", "##end", "b 1 1", "x-y 2 2"), ErrRoomName},
		{"room name is empty", file("1", "##start", "a 0 0", "##end", "b 1 1", " 2 2"), ErrRoomName},
		{"x is not an integer", file("1", "##start", "a 0 0", "##end", "b 1 1", "c x 2"), ErrCoordinates},
		{"y is not an integer", file("1", "##start", "a 0 0", "##end", "b 1 1", "c 1 y"), ErrCoordinates},
		{"decimal coordinate", file("1", "##start", "a 0 0", "##end", "b 1 1", "c 1.5 2"), ErrCoordinates},

		// Tunnels.
		{"tunnel to an unknown room", file("1", "##start", "a 0 0", "##end", "b 1 1", "a-zzz"), ErrUnknownRoom},
		{"room linked to itself", file("1", "##start", "a 0 0", "##end", "b 1 1", "a-a"), ErrSelfLink},
		{"same tunnel twice", file("1", "##start", "a 0 0", "##end", "b 1 1", "a-b", "a-b"), ErrDuplicateLink},
		{"same tunnel written both ways", file("1", "##start", "a 0 0", "##end", "b 1 1", "a-b", "b-a"), ErrDuplicateLink},
		{"audit badexample01 (self-link, duplicated tunnel)", auditBadExample01, ErrSelfLink},

		// Lines that are nothing.
		{"a word", file("1", "##start", "a 0 0", "##end", "b 1 1", "garbage"), ErrLine},
		{"a room with two fields", file("1", "##start", "a 0 0", "##end", "b 1 1", "c 1"), ErrLine},
		{"a room with a double space", file("1", "##start", "a 0 0", "##end", "b 1 1", "c  1 2"), ErrLine},
		{"a blank line inside the data", file("1", "##start", "a 0 0", "", "##end", "b 1 1", "a-b"), ErrLine},
		{"a blank line right after ##start", file("1", "##start", "", "a 0 0", "##end", "b 1 1", "a-b"), ErrLine},
		{"a line of spaces inside the data", file("1", "##start", "a 0 0", "   ", "##end", "b 1 1", "a-b"), ErrLine},
		{"tunnel with no second room", file("1", "##start", "a 0 0", "##end", "b 1 1", "a-"), ErrLine},
		{"tunnel with no first room", file("1", "##start", "a 0 0", "##end", "b 1 1", "-b"), ErrLine},
		{"tunnel with two dashes", file("1", "##start", "a 0 0", "##end", "b 1 1", "a-b-c"), ErrLine},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := Parse(strings.NewReader(tc.input))
			if !errors.Is(err, tc.want) {
				t.Fatalf("got error %v, want %v", err, tc.want)
			}
		})
	}
}

// TestParseNeverPanics feeds the parser garbage. Every input here is
// invalid: the test checks for an error and for the absence of a crash.
func TestParseNeverPanics(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"only newlines", "\n\n\n\n"},
		{"only comments", "#one\n#two\n##three\n"},
		{"binary bytes", string([]byte{0, 1, 2, 255, 254, 10, 0, 13, 10, 200})},
		{"a single NUL", "\x00"},
		{"a line of 2 MB", strings.Repeat("a", 2<<20)},
		{"a line of 10000 dashes", file("1", "##start", "a 0 0", "##end", "b 1 1", strings.Repeat("-", 10000))},
		{"a line of 10000 spaces", strings.Repeat(" ", 10000)},
		{"many fields in a room", file("1", "##start", strings.Repeat("a ", 5000))},
		{"only commands", "##start\n##end\n##start\n"},
		{"ant count, then nothing", "5\n"},
		{"huge ant count digits", strings.Repeat("9", 5000) + "\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := Parse(strings.NewReader(tc.input)); err == nil {
				t.Fatal("expected an error, got none")
			}
		})
	}
}

// ParseFile touches the real file system, so this test uses real files.
func TestParseFile(t *testing.T) {
	// t.TempDir creates an empty folder just for this test and deletes it
	// when the test ends, so we leave nothing behind on the disk.
	dir := t.TempDir()

	good := filepath.Join(dir, "good.txt")
	if err := os.WriteFile(good, []byte(withAnts("2")), 0o644); err != nil {
		t.Fatal(err)
	}
	colony, _, err := ParseFile(good)
	if err != nil {
		t.Fatalf("ParseFile(good) returned an error: %v", err)
	}
	if colony.Ants != 2 {
		t.Fatalf("Ants = %d, want 2", colony.Ants)
	}

	bad := filepath.Join(dir, "bad.txt")
	if err := os.WriteFile(bad, []byte(withAnts("0")), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ParseFile(bad); !errors.Is(err, ErrAnts) {
		t.Fatalf("ParseFile(bad) error = %v, want %v", err, ErrAnts)
	}

	// Two ways to fail before any parsing: a file that does not exist, and a
	// path that is a directory (opening it works, reading it fails).
	if _, _, err := ParseFile(filepath.Join(dir, "missing.txt")); !errors.Is(err, ErrRead) {
		t.Fatalf("ParseFile(missing) error = %v, want %v", err, ErrRead)
	}
	if _, _, err := ParseFile(dir); !errors.Is(err, ErrRead) {
		t.Fatalf("ParseFile(directory) error = %v, want %v", err, ErrRead)
	}
}

// FuzzParse checks that Parse never panics, and that whatever it accepts is a
// consistent colony. Run it with: go test ./internal/parser -fuzz=FuzzParse
//
// Fuzzing means Go invents thousands of random inputs, starting from the
// examples we give with f.Add, and mutates them to look for a crash. We cannot
// say what the right answer is for random input, so we check INVARIANTS:
// things that must be true for every colony Parse accepts.
// A normal "go test" only runs the examples, as a quick regression test.
func FuzzParse(f *testing.F) {
	// The seed inputs: a few good and a few strange starting points.
	f.Add([]byte(file(subjectLines...)))
	f.Add([]byte(withAnts("3")))
	f.Add([]byte(""))
	f.Add([]byte("1\n##start\na 0 0\n##end\nb 1 1\na-b\n"))
	f.Add([]byte("\x00\xff\n##start\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		colony, lines, err := Parse(bytes.NewReader(data))
		if err != nil {
			return // rejecting an input is always allowed; only a panic would fail
		}

		// From here the input was accepted, so the colony must make sense.
		if colony.Ants <= 0 {
			t.Fatalf("accepted a colony with %d ants", colony.Ants)
		}
		if len(lines) == 0 {
			t.Fatal("accepted a colony but returned no lines")
		}
		if colony.Start == "" || colony.End == "" || colony.Start == colony.End {
			t.Fatalf("bad start/end: %q, %q", colony.Start, colony.End)
		}
		names := make(map[string]bool)
		for _, room := range colony.Rooms {
			if names[room.Name] {
				t.Fatalf("duplicated room %q was accepted", room.Name)
			}
			names[room.Name] = true
		}
		if !names[colony.Start] || !names[colony.End] {
			t.Fatal("start or end is not in the list of rooms")
		}
		for _, link := range colony.Links {
			if !names[link[0]] || !names[link[1]] || link[0] == link[1] {
				t.Fatalf("bad tunnel %v was accepted", link)
			}
		}
	})
}
