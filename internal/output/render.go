// Package output turns the result of the simulation into the text the program
// prints. It only formats: it decides nothing and computes nothing.
package output

import (
	"fmt"
	"strings"

	"lem-in/internal/sim"
)

// Render returns the text that the program must print (section 4 of the PRD):
//
//  1. the lines of the input file, exactly as the parser gave them (comments
//     included);
//  2. one empty line;
//  3. one line for each turn, with its moves written "L<ant>-<room>" and
//     separated by one space, with no space at the end of the line.
//
// Every line, the last one too, ends with a newline. That is how a text file
// is normally written, and it keeps the prompt of the terminal on its own line
// after the program ends.
//
// Render does not sort, check or fix anything. It prints the turns it gets, in
// the order it gets them. If the moves of a turn are out of order, the mistake
// is in the code that produced them (sim.Run), and hiding it here would only
// hide it from the tests of that code.
func Render(lines []string, turns []sim.Turn) string {
	// A Builder collects text in one buffer that grows as needed. Writing to
	// it is cheap. Joining strings with "+=" in a loop is not: a string cannot
	// be changed, so every "+=" copies all the text written so far.
	var b strings.Builder

	for _, line := range lines {
		b.WriteString(line)
		b.WriteByte('\n')
	}

	// The parser already removed the empty lines at the end of the file, so
	// the one we write here is the only empty line between the two parts.
	b.WriteByte('\n')

	for _, turn := range turns {
		for i, move := range turn {
			// The space goes BEFORE each move except the first, so there is
			// never a space at the end of the line.
			if i > 0 {
				b.WriteByte(' ')
			}
			// A Builder can be written to like a file, so Fprintf puts the
			// formatted move straight into it, without a temporary string.
			fmt.Fprintf(&b, "L%d-%s", move.Ant, move.Room)
		}
		b.WriteByte('\n')
	}
	return b.String()
}
