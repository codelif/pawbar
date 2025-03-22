package modules

import (
	"github.com/gdamore/tcell/v2"
)

var (
	URGENT  = tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorBlack)
	ACTIVE  = tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
	SPECIAL = tcell.StyleDefault.Background(tcell.ColorDarkGreen).Foreground(tcell.ColorWhite)
	DEFAULT = tcell.StyleDefault
)

type EventCell struct {
	C        rune
	Style    tcell.Style
	Metadata string
	Mod        Module
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
	Cell EventCell
	TcellEvent tcell.Event
}

