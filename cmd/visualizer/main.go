// Command visualizer shows the ants of lem-in moving through the colony.
//
// Usage:
//
//	./lem-in ant-farm.txt | ./visualizer [-delay 300ms] [-width 78] [-height 20] [-color=false]
//	./lem-in ant-farm.txt | ./visualizer -html colony.html
//
// It reads what lem-in prints on its standard input. The first block is the
// colony and the second the moves. It draws one frame per turn, and redraws
// the screen between frames with ANSI escape codes.
//
// With -html it writes an HTML page instead. Open the page in a browser to
// watch the ants walk, with play, pause, step and a speed slider.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

// clearScreen moves the cursor to the top left and erases the screen, so each
// frame is drawn over the previous one.
const clearScreen = "\x1b[H\x1b[2J"

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		// Like lem-in, the visualizer prints the problem and returns normally.
		var solver *solverError
		if errors.As(err, &solver) {
			fmt.Println(solver.Line)
			return
		}
		fmt.Println("ERROR:", err)
	}
}

// run is the whole program. It reads the input from in and writes the frames
// to out, so a test can use text in memory instead of the real terminal.
func run(args []string, in io.Reader, out io.Writer) error {
	flags := flag.NewFlagSet("visualizer", flag.ContinueOnError)
	// The flag package prints its own message on a mistake. main prints the
	// error instead, so the package is kept quiet to avoid saying it twice.
	flags.SetOutput(io.Discard)
	delay := flags.Duration("delay", 500*time.Millisecond, "pause between two frames")
	width := flags.Int("width", 78, "width of the drawing in characters")
	height := flags.Int("height", 20, "height of the drawing in rows")
	colour := flags.Bool("color", true, "use colours")
	htmlPath := flags.String("html", "", "write an animated HTML page to this file instead of drawing in the terminal")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q (the colony is read from the standard input)", flags.Arg(0))
	}
	if *delay < 0 {
		return fmt.Errorf("the delay must not be negative, got %v", *delay)
	}

	data, err := readInput(in)
	if err != nil {
		return err
	}
	if *htmlPath != "" {
		return saveHTML(*htmlPath, data, out)
	}
	l, err := newLayout(data.Colony.Rooms, *width, *height)
	if err != nil {
		return err
	}

	// Frame 0 shows the colony before the first turn: all ants in the start
	// room. Then one frame per turn.
	current := newState(data.Colony.Ants, data.Colony.Start)
	total := len(data.Turns)
	show := func(turn int) error {
		_, err := fmt.Fprintf(out, "%sTurn %d/%d\n%s", clearScreen, turn, total,
			renderFrame(data.Colony, l, current.occupancy(), *colour))
		return err
	}

	if err := show(0); err != nil {
		return err
	}
	for i, turn := range data.Turns {
		time.Sleep(*delay)
		if err := current.apply(turn); err != nil {
			return fmt.Errorf("turn %d: %w", i+1, err)
		}
		if err := show(i + 1); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(out, "Done: %d ants reached the end in %d turns.\n", data.Colony.Ants, total)
	return err
}

// saveHTML builds the page in memory first and writes the file only when the
// page is complete, so a bad input never leaves a half written file behind.
func saveHTML(path string, data input, out io.Writer) error {
	var page bytes.Buffer
	if err := writeHTML(&page, data); err != nil {
		return err
	}
	if err := os.WriteFile(path, page.Bytes(), 0o644); err != nil {
		return fmt.Errorf("cannot write the page: %w", err)
	}
	_, err := fmt.Fprintf(out, "Wrote %s. Open it in a browser to watch the ants.\n", path)
	return err
}
