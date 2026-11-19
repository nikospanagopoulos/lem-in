package main

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"lem-in/internal/graph"
	"lem-in/internal/parser"
	"lem-in/internal/paths"
	"lem-in/internal/sim"
)

func TestNewStateAllAntsAtStart(t *testing.T) {
	s := newState(3, "start")
	want := map[string][]int{"start": {1, 2, 3}}
	if got := s.occupancy(); !reflect.DeepEqual(got, want) {
		t.Fatalf("occupancy = %v, want %v", got, want)
	}
}

// The first turns of Example 1 of the subject, one by one.
func TestApplySubjectExample(t *testing.T) {
	s := newState(3, "1")

	steps := []struct {
		turn sim.Turn
		want map[string][]int
	}{
		{
			sim.Turn{{Ant: 1, Room: "3"}, {Ant: 2, Room: "2"}},
			map[string][]int{"1": {3}, "3": {1}, "2": {2}},
		},
		{
			sim.Turn{{Ant: 1, Room: "4"}, {Ant: 2, Room: "5"}, {Ant: 3, Room: "3"}},
			map[string][]int{"4": {1}, "5": {2}, "3": {3}},
		},
		{
			sim.Turn{{Ant: 1, Room: "0"}, {Ant: 2, Room: "6"}, {Ant: 3, Room: "4"}},
			map[string][]int{"0": {1}, "6": {2}, "4": {3}},
		},
		{
			sim.Turn{{Ant: 2, Room: "0"}, {Ant: 3, Room: "0"}},
			map[string][]int{"0": {1, 2, 3}},
		},
	}
	for n, step := range steps {
		if err := s.apply(step.turn); err != nil {
			t.Fatalf("turn %d: %v", n+1, err)
		}
		if got := s.occupancy(); !reflect.DeepEqual(got, step.want) {
			t.Errorf("after turn %d: occupancy = %v, want %v", n+1, got, step.want)
		}
	}
}

// An ant that does not move in a turn stays where it was.
func TestApplyKeepsAntsThatDoNotMove(t *testing.T) {
	s := newState(2, "s")
	if err := s.apply(sim.Turn{{Ant: 1, Room: "a"}}); err != nil {
		t.Fatal(err)
	}
	want := map[string][]int{"a": {1}, "s": {2}}
	if got := s.occupancy(); !reflect.DeepEqual(got, want) {
		t.Fatalf("occupancy = %v, want %v", got, want)
	}
}

// The numbers of the ants of a room come out in increasing order, whatever the
// order of the moves.
func TestOccupancyIsSorted(t *testing.T) {
	s := newState(4, "s")
	if err := s.apply(sim.Turn{{Ant: 4, Room: "a"}, {Ant: 2, Room: "a"}, {Ant: 3, Room: "a"}}); err != nil {
		t.Fatal(err)
	}
	if got, want := s.occupancy()["a"], []int{2, 3, 4}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ants in a = %v, want %v", got, want)
	}
}

// A move of an ant that does not exist is an error, and the state is not
// changed, not even by the good moves that come before it in the same turn.
func TestApplyUnknownAnt(t *testing.T) {
	for _, ant := range []int{0, -1, 4} {
		s := newState(3, "s")
		err := s.apply(sim.Turn{{Ant: 1, Room: "a"}, {Ant: ant, Room: "b"}})
		if !errors.Is(err, errUnknownAnt) {
			t.Fatalf("ant %d: error = %v, want %v", ant, err, errUnknownAnt)
		}
		want := map[string][]int{"s": {1, 2, 3}}
		if got := s.occupancy(); !reflect.DeepEqual(got, want) {
			t.Errorf("ant %d: the state changed to %v", ant, got)
		}
	}
}

// With the real solver, on every good file of testdata/: the visualizer's
// state must follow the ants to the end room, one turn at a time, and in the
// end every ant must be there.
func TestStateEndsWithAllAntsInTheEndRoom(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "..", "testdata", "*.txt"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no test files found: %v", err)
	}
	checked := 0
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), "bad") {
			continue // these files have no solution
		}
		colony, _, err := parser.ParseFile(file)
		if err != nil {
			t.Fatalf("%s: %v", file, err)
		}
		sets, err := paths.Sets(graph.New(colony), colony.Start, colony.End)
		if err != nil {
			t.Fatalf("%s: %v", file, err)
		}
		turns := sim.Run(paths.Best(sets, colony.Ants), colony.Ants)

		s := newState(colony.Ants, colony.Start)
		for n, turn := range turns {
			if err := s.apply(turn); err != nil {
				t.Fatalf("%s turn %d: %v", file, n+1, err)
			}
		}
		if got := len(s.occupancy()[colony.End]); got != colony.Ants {
			t.Errorf("%s: %d ants in the end room, want %d", file, got, colony.Ants)
		}
		checked++
	}
	if checked < 15 {
		t.Errorf("only %d files were checked", checked)
	}
}
