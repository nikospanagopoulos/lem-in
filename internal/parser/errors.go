package parser

import "errors"

// The error values returned by the parser.
//
// Each one is created once, here, with errors.New. A caller recognises an error
// by identity, not by its text: errors.Is(err, parser.ErrAnts). That way we can
// change a message later without breaking any code that checks for the error.
//
// The text of each one is the reason that follows
// "ERROR: invalid data format, " in the program's output.
var (
	ErrAnts              = errors.New("invalid number of Ants")
	ErrNoStart           = errors.New("no start room found")
	ErrNoEnd             = errors.New("no end room found")
	ErrDuplicateStartEnd = errors.New("duplicated start or end room")
	ErrDuplicateRoom     = errors.New("duplicated room")
	ErrRoomName          = errors.New("invalid room name")
	ErrCoordinates       = errors.New("invalid coordinates")
	ErrUnknownRoom       = errors.New("link to unknown room")
	ErrSelfLink          = errors.New("room linked to itself")
	ErrDuplicateLink     = errors.New("duplicated link")
	ErrLine              = errors.New("invalid line")
	ErrRead              = errors.New("cannot read file")
)
