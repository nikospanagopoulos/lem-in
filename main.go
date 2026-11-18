// Command lem-in reads a colony from a file and prints how its ants reach the
// end room in the fewest turns.
//
// This file only wires the packages together. It holds no algorithm: parsing
// is in internal/parser, path-finding in internal/paths, the movement of the
// ants in internal/sim and the text format in internal/output.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"lem-in/internal/graph"
	"lem-in/internal/output"
	"lem-in/internal/parser"
	"lem-in/internal/paths"
	"lem-in/internal/sim"
)

// reasons maps each error value to the text that follows
// "ERROR: invalid data format, " (PRD section 9). It is a slice and not a map
// on purpose: the order is fixed, and a slice needs no lookup by key because
// errors.Is does the matching.
var reasons = []struct {
	err    error
	reason string
}{
	{parser.ErrAnts, "invalid number of Ants"},
	{parser.ErrNoStart, "no start room found"},
	{parser.ErrNoEnd, "no end room found"},
	{parser.ErrDuplicateStartEnd, "duplicated start or end room"},
	{parser.ErrDuplicateRoom, "duplicated room"},
	{parser.ErrRoomName, "invalid room name"},
	{parser.ErrCoordinates, "invalid coordinates"},
	{parser.ErrUnknownRoom, "link to unknown room"},
	{parser.ErrSelfLink, "room linked to itself"},
	{parser.ErrDuplicateLink, "duplicated link"},
	{parser.ErrLine, "invalid line"},
	{paths.ErrNoPath, "no path between start and end"},
	{parser.ErrRead, "cannot read file"},
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		// Only main prints an error. There is no os.Exit(1) on purpose: with
		// "go run ." a non-zero exit status adds an extra "exit status 1"
		// line that the audit does not expect.
		fmt.Println(message(err))
	}
}

// run is the whole program. It writes the result to w and returns an error
// instead of printing it, so that a test can call it with a buffer and read
// what it wrote. Nothing is written to w unless everything succeeded: the
// output is built as one string first, so a failure never leaves half a
// result behind.
func run(args []string, w io.Writer) error {
	if len(args) != 1 {
		// Wrong number of arguments is reported as "cannot read file", as in
		// the table of the PRD. %w keeps ErrRead recognisable by errors.Is.
		return fmt.Errorf("%w: expected one file argument, got %d", parser.ErrRead, len(args))
	}

	colony, lines, err := parser.ParseFile(args[0])
	if err != nil {
		return err
	}

	sets, err := paths.Sets(graph.New(colony), colony.Start, colony.End)
	if err != nil {
		return err
	}
	best := paths.Best(sets, colony.Ants)
	turns := sim.Run(best, colony.Ants)

	_, err = io.WriteString(w, output.Render(lines, turns))
	return err
}

// message returns the single line that main prints for err. errors.Is looks
// through every layer added with %w, so a wrapped error is still recognised.
// An error that is not in the table gets the plain message, with no reason.
func message(err error) string {
	for _, entry := range reasons {
		if errors.Is(err, entry.err) {
			return "ERROR: invalid data format, " + entry.reason
		}
	}
	return "ERROR: invalid data format"
}
