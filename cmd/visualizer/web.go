package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// The browser version. Instead of redrawing a terminal, the visualizer writes
// one HTML file that contains the colony, the turns and a small program
// (JavaScript) that animates them. The file needs no server and no internet:
// it opens by double click.
//
// The page is made of three files in the web folder. //go:embed copies them
// into the binary when it is built, so ./visualizer does not need the folder
// at run time.

//go:embed web/index.html web/app.js web/style.css
var assets embed.FS

// errUnknownRoom is returned when a move names a room that is not in the colony.
var errUnknownRoom = errors.New("unknown room")

// The JSON sent to the page. Short field names keep the file small: example07
// has about 500 turns of up to 100 moves each.
type pageMove struct {
	Ant  int    `json:"a"`
	Room string `json:"r"`
}

type pageRoom struct {
	Name string `json:"name"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

type pageData struct {
	Ants  int          `json:"ants"`
	Start string       `json:"start"`
	End   string       `json:"end"`
	Rooms []pageRoom   `json:"rooms"`
	Links [][2]string  `json:"links"`
	Turns [][]pageMove `json:"turns"`
}

// writeHTML writes the whole page to w.
func writeHTML(w io.Writer, data input) error {
	page, err := newPageData(data)
	if err != nil {
		return err
	}
	// json.Marshal escapes <, > and & (as < ...), so a room name can
	// never close the <script> element it is written into.
	payload, err := json.Marshal(page)
	if err != nil {
		return fmt.Errorf("cannot encode the colony: %w", err)
	}

	index, err := assets.ReadFile("web/index.html")
	if err != nil {
		return err
	}
	style, err := assets.ReadFile("web/style.css")
	if err != nil {
		return err
	}
	script, err := assets.ReadFile("web/app.js")
	if err != nil {
		return err
	}

	// A Replacer makes one pass over the template. Text that is inserted is
	// not searched again, so a room named {{SCRIPT}} stays a room name.
	fill := strings.NewReplacer(
		"{{STYLE}}", string(style),
		"{{DATA}}", string(payload),
		"{{SCRIPT}}", string(script),
	)
	_, err = io.WriteString(w, fill.Replace(string(index)))
	return err
}

// newPageData converts the input to the JSON structure. It also plays every
// turn on a state, so a bad input is reported here, in Go, and not as a blank
// page in the browser.
func newPageData(data input) (pageData, error) {
	c := data.Colony
	known := make(map[string]bool, len(c.Rooms))
	rooms := make([]pageRoom, len(c.Rooms))
	for i, room := range c.Rooms {
		known[room.Name] = true
		rooms[i] = pageRoom{Name: room.Name, X: room.X, Y: room.Y}
	}

	current := newState(c.Ants, c.Start)
	turns := make([][]pageMove, len(data.Turns))
	for i, turn := range data.Turns {
		if err := current.apply(turn); err != nil {
			return pageData{}, fmt.Errorf("turn %d: %w", i+1, err)
		}
		moves := make([]pageMove, len(turn))
		for j, move := range turn {
			if !known[move.Room] {
				return pageData{}, fmt.Errorf("turn %d: %w %q", i+1, errUnknownRoom, move.Room)
			}
			moves[j] = pageMove{Ant: move.Ant, Room: move.Room}
		}
		turns[i] = moves
	}

	links := c.Links
	if links == nil {
		links = [][2]string{} // JSON null would break the page, [] does not
	}
	return pageData{
		Ants: c.Ants, Start: c.Start, End: c.End,
		Rooms: rooms, Links: links, Turns: turns,
	}, nil
}
