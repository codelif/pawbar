package main

import "github.com/gdamore/tcell/v2"

type EventCell struct {
	c        rune
	style    tcell.Style
	m        Module
	metadata string
}

type Module interface {
	Render() []EventCell
	Run() (<-chan bool, chan<- Event, error)
	Channels() (<-chan bool, chan<- Event)
  Name() string
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

func StartModules() (chan Module, []Module, []Module) {
	tleft, tright := modules()
	modev := make(chan Module)

	var left []Module
	var right []Module

	for _, mod := range append(tleft) {
		rec, _, err := mod.Run()
		if err == nil {
			go modevloop(mod, rec, modev)
			left = append(left, mod)
		}
	}

	for _, mod := range append(tright) {
		rec, _, err := mod.Run()
		if err == nil {
			go modevloop(mod, rec, modev)
			right = append(right, mod)
		}
	}

	return modev, left, right
}
