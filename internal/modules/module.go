package modules

import (
	"git.sr.ht/~rockorager/vaxis"
)

var (
// vaxis.StyleDefault
// DEFAULT = tcell.StyleDefault
// URGENT  = DEFAULT.Foreground(tcell.ColorRed)
// WARNING = DEFAULT.Foreground(tcell.ColorYellow)
// GOOD    = DEFAULT.Foreground(tcell.ColorGreen)
// ACTIVE  = DEFAULT.Foreground(tcell.ColorWhite)
// COOL    = DEFAULT.Foreground(tcell.ColorLightBlue)
// SPECIAL = DEFAULT.Foreground(tcell.ColorDarkGreen).Background(tcell.ColorWhite)
)

type EventCell struct {
	C        vaxis.Cell
	Metadata string
	Mod      Module
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

var (
	ECSPACE = EventCell{
		C: vaxis.Cell{Character: vaxis.Character{
			Grapheme: " ",
			Width:    1,
		}},
		Metadata: "",
		Mod:      nil,
	}
	ECDOT = EventCell{
		C: vaxis.Cell{Character: vaxis.Character{
			Grapheme: ".",
			Width:    1,
		}},
		Metadata: "",
		Mod:      nil,
	}
)
