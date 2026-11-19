package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// htmlOf plays the lem-in output text through the same path as the program:
// readInput, then writeHTML.
func htmlOf(t *testing.T, text string) string {
	t.Helper()
	data, err := readInput(strings.NewReader(text))
	if err != nil {
		t.Fatal(err)
	}
	var page strings.Builder
	if err := writeHTML(&page, data); err != nil {
		t.Fatal(err)
	}
	return page.String()
}

// The page is complete: every placeholder is replaced and the three embedded
// files are inside.
func TestWriteHTMLIsOneCompletePage(t *testing.T) {
	page := htmlOf(t, subjectOutput)

	for _, placeholder := range []string{"{{STYLE}}", "{{DATA}}", "{{SCRIPT}}"} {
		if strings.Contains(page, placeholder) {
			t.Errorf("the page still contains %s", placeholder)
		}
	}
	for _, want := range []string{
		"<!DOCTYPE html>",
		".room.start",          // from style.css
		"function buildFrames", // from app.js
		`id="colony"`,          // the data element
		`"ants":3`,             // the data itself
		`"start":"1","end":"0"`,
		`{"a":1,"r":"3"}`, // a move of the first turn
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the page does not contain %q", want)
		}
	}
	// No link to another file or to the internet: the page works on its own.
	for _, bad := range []string{"src=\"http", "href=\"http", "<link ", "src=\"app.js"} {
		if strings.Contains(page, bad) {
			t.Errorf("the page depends on something outside itself: %q", bad)
		}
	}
}

// A room name is data. It must not be able to end the <script> element or to
// be taken for a placeholder.
func TestWriteHTMLKeepsRoomNamesAsData(t *testing.T) {
	const colony = "1\n##start\n{{SCRIPT}} 0 0\n##end\n</script> 1 0\n{{SCRIPT}}-</script>\n\nL1-</script>\n"
	page := htmlOf(t, colony)

	// The only </script> in the page are the two that close its own scripts.
	if n := strings.Count(page, "</script>"); n != 2 {
		t.Errorf("%d closing script tags, want 2", n)
	}
	if !strings.Contains(page, `</script>`) {
		t.Error("the room name was not escaped")
	}
	// The name {{SCRIPT}} was not replaced by the script a second time.
	if n := strings.Count(page, "function buildFrames"); n != 1 {
		t.Errorf("app.js is in the page %d times, want 1", n)
	}
}

// A colony without tunnels must give [] and not null, which would break the
// page (null has no length).
func TestNewPageDataEmptyLinksIsAnArray(t *testing.T) {
	page, err := newPageData(input{})
	if err != nil {
		t.Fatal(err)
	}
	if page.Links == nil {
		t.Error("Links is nil, it would be written as null")
	}
}

// Bad moves are reported by Go, not by a blank page in the browser.
func TestWriteHTMLRejectsBadMoves(t *testing.T) {
	const head = "1\n##start\ns 0 0\n##end\ne 1 0\ns-e\n\n"
	tests := []struct {
		name  string
		moves string
		want  error
	}{
		{"unknown ant", "L2-e\n", errUnknownAnt},
		{"unknown room", "L1-nowhere\n", errUnknownRoom},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := readInput(strings.NewReader(head + tc.moves))
			if err != nil {
				t.Fatal(err)
			}
			err = writeHTML(&strings.Builder{}, data)
			if !errors.Is(err, tc.want) {
				t.Errorf("got %v, want %v", err, tc.want)
			}
		})
	}
}

// -html writes the file and says so. Nothing is drawn in the terminal.
func TestRunHTMLWritesTheFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "colony.html")

	var out strings.Builder
	if err := run([]string{"-html", path}, strings.NewReader(subjectOutput), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Wrote "+path) {
		t.Errorf("unexpected message %q", out.String())
	}
	if strings.Contains(out.String(), clearScreen) {
		t.Error("the terminal was drawn although -html was given")
	}
	file, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(file), "<!DOCTYPE html>") {
		t.Error("the file is not an HTML page")
	}
}

// When the page cannot be written the error says so, and a bad input leaves
// no file behind.
func TestRunHTMLErrors(t *testing.T) {
	dir := t.TempDir()

	missing := filepath.Join(dir, "no-such-folder", "colony.html")
	err := run([]string{"-html", missing}, strings.NewReader(subjectOutput), &strings.Builder{})
	if err == nil || !strings.Contains(err.Error(), "cannot write the page") {
		t.Errorf("got %v, want a 'cannot write the page' error", err)
	}

	bad := filepath.Join(dir, "bad.html")
	const colony = "1\n##start\ns 0 0\n##end\ne 1 0\ns-e\n\nL9-e\n"
	err = run([]string{"-html", bad}, strings.NewReader(colony), &strings.Builder{})
	if !errors.Is(err, errUnknownAnt) {
		t.Errorf("got %v, want errUnknownAnt", err)
	}
	if _, statErr := os.Stat(bad); !errors.Is(statErr, os.ErrNotExist) {
		t.Error("a file was written for a bad input")
	}
}
