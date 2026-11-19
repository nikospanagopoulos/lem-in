package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"lem-in/internal/graph"
	"lem-in/internal/parser"
	"lem-in/internal/sim"
)

// The visualizer reads what lem-in prints:
//
//	the colony, exactly as in the input file   <- block 1
//	(one empty line)
//	L1-3 L2-2                                  <- block 2: one line per turn
//	L1-4 L2-5 L3-3
//	...
//
// This file turns that text back into values the rest of the program can use.
// It draws nothing and prints nothing, so it is easy to test.

// input is everything the visualizer needs from the standard input.
type input struct {
	Colony graph.Colony
	Turns  []sim.Turn
}

// solverError is returned when lem-in itself printed an error instead of a
// result ("ERROR: invalid data format, ..."). Its text is that line, so main
// can show it unchanged. A struct pointer, and not errors.New, lets main tell
// this case apart from a problem in the visualizer's own input with errors.As.
type solverError struct{ Line string }

func (e *solverError) Error() string { return e.Line }

// The other ways the input can be wrong.
var (
	errEmptyInput  = errors.New("empty input")
	errNoSeparator = errors.New("no empty line between the colony and the turns")
	errBadMove     = errors.New("bad move")
)

// readInput reads the whole text from r and returns the colony and the turns.
func readInput(r io.Reader) (input, error) {
	lines, err := readLines(r)
	if err != nil {
		return input{}, err
	}
	if len(lines) == 0 {
		return input{}, errEmptyInput
	}
	// When lem-in rejects a file it prints one ERROR line and nothing else.
	if strings.HasPrefix(lines[0], "ERROR") {
		return input{}, &solverError{Line: lines[0]}
	}

	// The one empty line splits the colony from the turns.
	blank := slices.Index(lines, "")
	if blank < 0 {
		return input{}, errNoSeparator
	}

	// Block 1 is a normal input file, so the parser of lem-in reads it. The
	// visualizer does not need a second parser. The parser also gives back the
	// lines, which we do not need here.
	colony, _, err := parser.Parse(strings.NewReader(strings.Join(lines[:blank], "\n")))
	if err != nil {
		// %w keeps the original error, so a caller can still use errors.Is.
		return input{}, fmt.Errorf("the colony in the input is not valid: %w", err)
	}

	turns, err := parseTurns(lines[blank+1:])
	if err != nil {
		return input{}, err
	}
	return input{Colony: colony, Turns: turns}, nil
}

// readLines returns the lines of r. The Scanner already removes the line
// ending, "\r\n" as well as "\n", so Windows files need no extra care.
func readLines(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	// A Scanner stops at a line longer than its buffer (64 KB by default). A
	// turn with 1000 ants is about 10 KB, but we leave a lot of room.
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("cannot read the input: %w", err)
	}
	return lines, nil
}

// parseTurns turns the lines of block 2 into turns. Empty lines at the end are
// ignored (a file may end with extra newlines); an empty line anywhere else
// means a turn without moves, which lem-in never prints, so it is an error.
func parseTurns(lines []string) ([]sim.Turn, error) {
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	turns := make([]sim.Turn, 0, len(lines))
	for i, line := range lines {
		words := strings.Fields(line)
		if len(words) == 0 {
			return nil, fmt.Errorf("%w: turn %d has no moves", errBadMove, i+1)
		}
		turn := make(sim.Turn, 0, len(words))
		for _, word := range words {
			move, err := parseMove(word)
			if err != nil {
				return nil, fmt.Errorf("turn %d: %w", i+1, err)
			}
			turn = append(turn, move)
		}
		turns = append(turns, turn)
	}
	return turns, nil
}

// parseMove reads one move such as "L12-roomname". A room name never contains
// a '-' (the parser of lem-in rejects it), so the first '-' is the separator.
func parseMove(word string) (sim.Move, error) {
	body, ok := strings.CutPrefix(word, "L")
	if !ok {
		return sim.Move{}, fmt.Errorf("%w: %q does not start with L", errBadMove, word)
	}
	number, room, ok := strings.Cut(body, "-")
	if !ok || room == "" {
		return sim.Move{}, fmt.Errorf("%w: %q is not of the form L<ant>-<room>", errBadMove, word)
	}
	ant, err := strconv.Atoi(number)
	if err != nil || ant < 1 {
		return sim.Move{}, fmt.Errorf("%w: %q has no valid ant number", errBadMove, word)
	}
	return sim.Move{Ant: ant, Room: room}, nil
}
