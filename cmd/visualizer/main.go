// Command visualizer shows the ants of lem-in moving through the colony.
//
// Usage: ./lem-in ant-farm.txt | ./visualizer
//
// TEMPORARY VERSION: it only reads the input and prints a summary. The
// drawing is added in the next steps; the reading code in input.go stays.
package main

import (
	"errors"
	"fmt"
	"os"
)

func main() {
	in, err := readInput(os.Stdin)
	if err != nil {
		// Like lem-in, the visualizer prints the problem and returns normally.
		var solver *solverError
		if errors.As(err, &solver) {
			fmt.Println(solver.Line)
			return
		}
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Printf("%d ants, %d rooms, %d tunnels, %d turns\n",
		in.Colony.Ants, len(in.Colony.Rooms), len(in.Colony.Links), len(in.Turns))
}
