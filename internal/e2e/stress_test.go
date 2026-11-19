package e2e

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"lem-in/internal/sim"
)

// Stress colony of PRD section 10.2: 3000 rooms, 9000 tunnels, 1000 ants.
const (
	stressRooms   = 3000
	stressTunnels = 9000
	stressAnts    = 1000
	stressLimit   = 2 * time.Second
)

// stressColony builds the text of a random colony. The seed is fixed, so every
// run builds the same colony and a failure can be repeated.
//
// A random graph may have no path from start to end. To be sure there is one,
// every room is first linked to a room that came before it: that makes one
// connected tree. The remaining tunnels are then added at random.
func stressColony() string {
	rng := rand.New(rand.NewSource(1))

	var b strings.Builder
	fmt.Fprintf(&b, "%d\n", stressAnts)
	for i := 0; i < stressRooms; i++ {
		switch i {
		case 0:
			b.WriteString("##start\n")
		case stressRooms - 1:
			b.WriteString("##end\n")
		}
		fmt.Fprintf(&b, "r%d %d %d\n", i, i%100, i/100)
	}

	// A set of the tunnels already written, so that no tunnel appears twice.
	// The key is always (smaller, bigger), so 3-5 and 5-3 count as the same.
	seen := make(map[[2]int]bool, stressTunnels)
	add := func(a, c int) bool {
		if a == c {
			return false
		}
		if a > c {
			a, c = c, a
		}
		if seen[[2]int{a, c}] {
			return false
		}
		seen[[2]int{a, c}] = true
		fmt.Fprintf(&b, "r%d-r%d\n", a, c)
		return true
	}

	for i := 1; i < stressRooms; i++ {
		add(i, rng.Intn(i))
	}
	for len(seen) < stressTunnels {
		add(rng.Intn(stressRooms), rng.Intn(stressRooms))
	}
	return b.String()
}

// writeStress writes the stress colony to a temporary file and returns its path.
func writeStress(tb testing.TB) string {
	tb.Helper()
	path := filepath.Join(tb.TempDir(), "stress.txt")
	if err := os.WriteFile(path, []byte(stressColony()), 0o644); err != nil {
		tb.Fatal(err)
	}
	return path
}

// The big colony must finish within the team target and give a valid answer.
func TestStress(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped with -short")
	}
	path := writeStress(t)

	start := time.Now()
	text := runProgram(t, path)
	elapsed := time.Since(start)
	t.Logf("%d rooms, %d tunnels, %d ants: %v", stressRooms, stressTunnels, stressAnts, elapsed)

	if elapsed > stressLimit {
		t.Errorf("took %v, the target is %v", elapsed, stressLimit)
	}

	_, turnLines := splitOutput(t, text)
	colony, _ := colonyOf(t, path)
	turns := parseMoves(t, turnLines)
	if err := sim.Check(colony, turns); err != nil {
		t.Errorf("the moves break a rule: %v", err)
	}
	t.Logf("%d turns", len(turns))
}

// BenchmarkProgram measures the whole program, started as a new process each
// time, on the two audit colonies with many ants and on the stress colony.
// Run it with: go test -bench . ./internal/e2e
func BenchmarkProgram(b *testing.B) {
	files := map[string]string{
		"example06": testdataPath("example06.txt"),
		"example07": testdataPath("example07.txt"),
		"stress":    writeStress(b),
	}
	for name, path := range files {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if err := exec.Command(programPath, path).Run(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
