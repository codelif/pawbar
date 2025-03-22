package main

import "github.com/gdamore/tcell/v2"

var (
	URGENT  = tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorBlack)
	ACTIVE  = tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
	SPECIAL = tcell.StyleDefault.Background(tcell.ColorDarkGreen).Foreground(tcell.ColorWhite)
	DEFAULT = tcell.StyleDefault
)

type EventCell struct {
	c        rune
	style    tcell.Style
	metadata string
	m        Module
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
	c EventCell
	e tcell.Event
}

func modevloop(mod Module, rec <-chan bool, modev chan<- Module) {
	for <-rec {
		logger.Printf("Module '%s' sent a Render event.\n", mod.Name())
		modev <- mod
	}
}

func RunServices(mods []Module) {
	neededServices := make(map[string]bool)
	for _, mod := range mods {
		for _, dep := range mod.Dependencies() {
			neededServices[dep] = true
		}
	}

	// 2) Start each needed service
	for dep := range neededServices {
		svc, ok := serviceRegistry[dep]
		if !ok {
			logger.Printf("Unknown service dependency: %s\n", dep)
			continue
		}
		svc.Start() // handle errors etc.
	}
}

func StartModules() (chan Module, []Module, []Module) {
	tleft, tright := modules()
  RunServices(append(tleft, tright...))
	modev := make(chan Module)

	var left []Module
	var right []Module

	for _, mod := range tleft {
		rec, _, err := mod.Run()
		if err == nil {
			go modevloop(mod, rec, modev)
			left = append(left, mod)
		}
	}

	for _, mod := range tright {
		rec, _, err := mod.Run()
		if err == nil {
			go modevloop(mod, rec, modev)
			right = append(right, mod)
		}
	}

	return modev, left, right
}
