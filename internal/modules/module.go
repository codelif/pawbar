package modules

import (
	"github.com/gdamore/tcell/v2"
)

var (
	DEFAULT = tcell.StyleDefault
	URGENT  = DEFAULT.Foreground(tcell.ColorRed)
	WARNING = DEFAULT.Foreground(tcell.ColorYellow)
	GOOD    = DEFAULT.Foreground(tcell.ColorGreen)
	ACTIVE  = DEFAULT.Foreground(tcell.ColorWhite)
	COOL    = DEFAULT.Foreground(tcell.ColorLightBlue)
	SPECIAL = DEFAULT.Foreground(tcell.ColorDarkGreen).Background(tcell.ColorWhite)
)

type EventCell struct {
	C        rune
	Style    tcell.Style
	Metadata string
	Mod      Module
}

type StaticCell struct {
	c     rune
	style tcell.Style
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
	TcellEvent tcell.Event
}
