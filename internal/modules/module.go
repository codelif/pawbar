package modules

import (
	"git.sr.ht/~rockorager/vaxis"
)

var (
	BLACK   = vaxis.IndexColor(0)
	URGENT  = vaxis.IndexColor(9)
	WARNING = vaxis.IndexColor(11)
	GOOD    = vaxis.IndexColor(2)
	ACTIVE  = vaxis.IndexColor(15)
	COOL    = vaxis.RGBColor(173, 216, 230)
	SPECIAL = vaxis.RGBColor(0, 100, 0)
)

func Cell(r rune, s vaxis.Style) vaxis.Cell {
	return vaxis.Cell{Character: vaxis.Characters(string(r))[0], Style: s}
}

type EventCell struct {
	C          vaxis.Cell
	Metadata   string
	Mod        Module
	MouseShape vaxis.MouseShape
}

type Module interface {
	Render() []EventCell
	Run() (<-chan bool, chan<- Event, error)
	Channels() (<-chan bool, chan<- Event)
	Name() string
	Dependencies() []string
}

type Event struct {
	Cell       EventCell
	VaxisEvent vaxis.Event
}

type FocusIn struct {
	NewMod  Module
	PrevMod Module
}
type FocusOut struct {
	NewMod  Module
	PrevMod Module
}

func (FocusIn) String() string  { return "FocusIn" }
func (FocusOut) String() string { return "FocusOut" }

var (
	ECSPACE = EventCell{
		C:          vaxis.Cell{Character: vaxis.Characters(" ")[0]},
		Metadata:   "",
		Mod:        nil,
		MouseShape: "",
	}
	ECDOT = EventCell{
		C:          vaxis.Cell{Character: vaxis.Characters(".")[0]},
		Metadata:   "",
		Mod:        nil,
		MouseShape: "",
	}
	ECELLIPSIS = EventCell{
		C:          vaxis.Cell{Character: vaxis.Characters("…")[0]},
		Metadata:   "",
		Mod:        nil,
		MouseShape: "",
	}
)
