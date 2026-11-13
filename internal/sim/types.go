package sim

// Move is one ant entering one room during one turn. In the program's output
// it is written as "L<Ant>-<Room>".
//
// A Move only says where the ant ARRIVES. It does not say where the ant came
// from: whoever reads a list of turns (like Check) has to follow each ant from
// turn to turn to know that.
type Move struct {
	Ant  int    // ant number, from 1 to the number of ants
	Room string // name of the room the ant enters
}

// Turn is everything that happens in one turn: all the moves, sorted by ant
// number. Only ants that moved appear in it.
type Turn []Move
