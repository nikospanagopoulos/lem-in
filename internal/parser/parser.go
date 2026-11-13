// Package parser reads an ant farm description and turns it into a
// graph.Colony, or into one of the errors of errors.go.
//
// Kinds of line, and what Parse does with each one (line 1 is always the
// number of ants):
//
//	3                a positive integer, first line only   -> number of ants
//	##start          exactly this text                     -> next room is the start
//	##end            exactly this text                     -> next room is the end
//	#anything        comment or unknown ## command          -> ignored
//	name x y         three fields, one space between them  -> a room
//	name1-name2      no space, one dash                    -> a tunnel
//	(empty line)                                           -> error
//	anything else                                          -> error
//
// Rooms and tunnels may come in any order. Everything that needs the whole file
// (the tunnels' rooms exist, start and end exist) is checked after the last line.
package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"lem-in/internal/graph"
)

// maxLineBytes is the longest line Parse accepts. A longer line is invalid.
const maxLineBytes = 1 << 20 // 1 MB

// pending says what the last ##start or ##end line is still waiting for.
// Those two commands apply to the NEXT room line, so while reading we must
// remember that one of them is "open".
type pending int

// iota counts up from 0 inside a const block: nothing = 0, wantStart = 1,
// wantEnd = 2. It is Go's way of writing a small list of named values.
const (
	nothing   pending = iota // no command is waiting for a room
	wantStart                // ##start was read, its room has not come yet
	wantEnd                  // ##end was read, its room has not come yet
)

// ParseFile opens the file at path and parses it with Parse.
func ParseFile(path string) (graph.Colony, []string, error) {
	f, err := os.Open(path)
	if err != nil {
		// %w wraps an error inside a new one. Wrapping two means that
		// errors.Is finds both: ErrRead (our value) and the original cause
		// from the operating system, which keeps the details for debugging.
		return graph.Colony{}, nil, fmt.Errorf("%w: %w", ErrRead, err)
	}
	// defer runs this line when the function returns, however it returns.
	// It guarantees the file is closed, with no need to repeat Close on
	// every return path.
	defer f.Close()

	// An open file is an io.Reader, so the same Parse works for files and for
	// text held in memory (which is how the tests call it).
	return Parse(f)
}

// Parse reads a colony from r. Besides the colony it returns the lines of the
// input exactly as they were read, comments included, minus the blank lines at
// the end of the file. The program prints those lines before the moves.
//
// When several things are wrong, the error returned is the first of: the ant
// count; a bad line, in file order; a ##start or ##end without a room; missing
// start; missing end; a bad tunnel, in file order.
func Parse(r io.Reader) (graph.Colony, []string, error) {
	// Two steps, kept apart on purpose: first get the lines (this can fail
	// because of the file), then understand them (this can fail because of the
	// content). Each step is small enough to test alone.
	lines, err := readLines(r)
	if err != nil {
		return graph.Colony{}, nil, err
	}
	colony, err := parseLines(lines)
	if err != nil {
		// On an error we return the zero Colony, never a half-built one, so a
		// caller that forgets to check err cannot use broken data by mistake.
		return graph.Colony{}, nil, err
	}
	return colony, lines, nil
}

// readLines returns the lines of r without their line ending (a "\r\n" ending
// is handled like "\n") and without the blank lines at the end.
func readLines(r io.Reader) ([]string, error) {
	// A Scanner splits its input into lines. It already removes the "\n" and a
	// "\r" just before it, which is why files written on Windows work too.
	scanner := bufio.NewScanner(r)

	// By default a Scanner stops with an error on lines longer than 64 KB.
	// We choose the limit ourselves (maxLineBytes) so that the behaviour is
	// the same for every input: a longer line is simply invalid.
	scanner.Buffer(make([]byte, 0, 64*1024), maxLineBytes)

	var lines []string
	for scanner.Scan() { // Scan moves to the next line; false at the end or on error
		lines = append(lines, scanner.Text())
	}

	// Scan returns false both at the normal end and on an error. Err tells
	// us which one it was: nil means a normal end.
	if err := scanner.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return nil, ErrLine // a line over maxLineBytes
		}
		return nil, fmt.Errorf("%w: %w", ErrRead, err) // for example, a directory
	}

	// Many files end with an extra empty line. It is not data, so we drop it.
	// Empty lines in the middle are kept: parseLines rejects them.
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines, nil
}

// parseLines turns the lines of a file into a colony.
func parseLines(lines []string) (graph.Colony, error) {
	// An empty file has no first line, so there is no number of ants.
	if len(lines) == 0 {
		return graph.Colony{}, ErrAnts
	}
	ants, err := parseAnts(lines[0])
	if err != nil {
		return graph.Colony{}, err
	}

	colony := graph.Colony{Ants: ants}

	// What the parser remembers while it moves from line to line:
	known := make(map[string]bool) // names of the rooms read so far (a set)
	var haveStart, haveEnd bool    // has the start / the end room been read?
	waiting := nothing             // is a ##start or ##end waiting for its room?

	// lines[1:] skips line 0, the number of ants, which we already handled.
	for _, line := range lines[1:] {
		// A switch with no value after the word "switch" is an if / else-if
		// chain: the first case that is true runs, and the others are skipped.
		// So the ORDER of the cases matters, and each case below assumes that
		// the ones above it did not match.
		switch {

		// "##start" must come before the general "#" case, because it also
		// starts with "#".
		case line == "##start":
			if waiting != nothing {
				// Two commands in a row: the first one never got its room.
				return graph.Colony{}, missingRoom(waiting)
			}
			if haveStart {
				return graph.Colony{}, ErrDuplicateStartEnd
			}
			waiting = wantStart

		case line == "##end":
			if waiting != nothing {
				return graph.Colony{}, missingRoom(waiting)
			}
			if haveEnd {
				return graph.Colony{}, ErrDuplicateStartEnd
			}
			waiting = wantEnd

		case strings.HasPrefix(line, "#"):
			// A comment, or a ## command we do not know. The subject says
			// unknown commands are ignored, so there is nothing to do.
			// (The line is still kept in the raw lines that Parse returns.)

		case line == "":
			// An empty line in the middle of the data. Without this case it
			// would reach the last case below and give a different error
			// when a ##start is waiting.
			return graph.Colony{}, ErrLine

		// A room line has spaces ("name x y"); a tunnel line never does.
		// That is the simplest way to tell them apart.
		case strings.Contains(line, " "):
			room, err := parseRoom(line)
			if err != nil {
				return graph.Colony{}, err
			}
			if known[room.Name] { // reading a missing map key gives false
				return graph.Colony{}, ErrDuplicateRoom
			}
			known[room.Name] = true
			colony.Rooms = append(colony.Rooms, room)

			// If a ##start or ##end was waiting, this is its room.
			switch waiting {
			case wantStart:
				colony.Start = room.Name
				haveStart = true
			case wantEnd:
				colony.End = room.Name
				haveEnd = true
			}
			waiting = nothing

		// Everything else is expected to be a tunnel ("a-b").
		default:
			// A tunnel where a room was expected: ##start or ##end had no
			// room after it.
			if waiting != nothing {
				return graph.Colony{}, missingRoom(waiting)
			}
			link, err := parseLink(line)
			if err != nil {
				return graph.Colony{}, err
			}
			// We only store the tunnel here. We cannot yet say whether its
			// rooms exist: they may come later in the file.
			colony.Links = append(colony.Links, link)
		}
	}

	// The checks below need the whole file, so they run after the loop.

	// The file ended right after a ##start or ##end.
	if waiting != nothing {
		return graph.Colony{}, missingRoom(waiting)
	}
	if !haveStart {
		return graph.Colony{}, ErrNoStart
	}
	if !haveEnd {
		return graph.Colony{}, ErrNoEnd
	}
	// Only now do we know every room, so only now can we check the tunnels.
	if err := checkLinks(colony.Links, known); err != nil {
		return graph.Colony{}, err
	}
	return colony, nil
}

// missingRoom is the error for a ##start or ##end that no room followed.
func missingRoom(waiting pending) error {
	if waiting == wantStart {
		return ErrNoStart
	}
	return ErrNoEnd
}

// parseAnts reads the first line: digits only, greater than zero, and small
// enough for an int. So "+3", " 3", "3 " and "3.5" are all rejected.
func parseAnts(line string) (int, error) {
	if line == "" {
		return 0, ErrAnts
	}

	// We check the characters ourselves because strconv.Atoi is too generous:
	// it accepts "+3" and "-3". Ranging over a string gives its characters.
	for _, c := range line {
		if c < '0' || c > '9' {
			return 0, ErrAnts
		}
	}

	// Atoi is still needed to turn the digits into a number. It also fails
	// when the number does not fit in an int (for example 30 digits long).
	n, err := strconv.Atoi(line)
	if err != nil || n <= 0 {
		return 0, ErrAnts
	}
	return n, nil
}

// parseRoom reads a line "name x y": three fields, one space between them.
// A name cannot be empty, start with L, or contain a dash (our own rule, so
// that a tunnel line can be cut at its dash). A name starting with # never gets
// here: such a line is a comment.
func parseRoom(line string) (graph.Room, error) {
	// strings.Split cuts at every single space and keeps empty pieces, so
	// "c  1 2" (two spaces) gives four fields and is rejected. strings.Fields
	// would hide the extra space and accept a line that is not in the format.
	fields := strings.Split(line, " ")
	if len(fields) != 3 {
		return graph.Room{}, ErrLine
	}

	name := fields[0]
	if name == "" || name[0] == 'L' || strings.Contains(name, "-") {
		return graph.Room{}, ErrRoomName
	}

	// Atoi returns the number and an error. We read both coordinates first and
	// look at the errors after, so one check covers x and y.
	x, errX := strconv.Atoi(fields[1])
	y, errY := strconv.Atoi(fields[2])
	if errX != nil || errY != nil {
		return graph.Room{}, ErrCoordinates
	}
	return graph.Room{Name: name, X: x, Y: y}, nil
}

// parseLink reads a line "a-b": exactly one dash, a name on each side. Whether
// the two rooms exist is checked later, by checkLinks.
func parseLink(line string) ([2]string, error) {
	// "a-b" gives two pieces. "a-b-c" gives three, "a-" and "-b" give an empty
	// piece, and a line with no dash at all gives one piece: all invalid.
	parts := strings.Split(line, "-")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return [2]string{}, ErrLine
	}
	return [2]string{parts[0], parts[1]}, nil
}

// checkLinks checks the tunnels, in file order: no room linked to itself, both
// rooms exist, and no tunnel appears twice ("a-b" and "b-a" are the same one).
func checkLinks(links [][2]string, known map[string]bool) error {
	// seen is a set of tunnels. An array like [2]string can be used as a map
	// key (a slice cannot), which is why a tunnel is stored as an array.
	seen := make(map[[2]string]bool, len(links))

	for _, link := range links {
		a, b := link[0], link[1]
		if a == b {
			return ErrSelfLink
		}
		if !known[a] || !known[b] {
			return ErrUnknownRoom
		}

		// "a-b" and "b-a" are the same tunnel, so before looking it up we put
		// its two names in a fixed (alphabetical) order. Both spellings then
		// give the same key.
		key := link
		if a > b {
			key = [2]string{b, a}
		}
		if seen[key] {
			return ErrDuplicateLink
		}
		seen[key] = true
	}
	return nil
}
