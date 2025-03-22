package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

func NewClock() Module {
	return &Clock{}
}

type Clock struct {
	receive chan bool
	send    chan Event
}

func (c *Clock) Run() (<-chan bool, chan<- Event, error) {
	c.receive = make(chan bool)
	c.send = make(chan Event)

	go func() {
		t := time.NewTicker(time.Second)
		for {
			select {
			case <-t.C:
				c.receive <- true
			case e := <-c.send:
        switch ev := e.e.(type) {
        case *tcell.EventMouse:
          if ev.Buttons() != 0{
            x, y := ev.Position()
            logger.Printf("Clock: Got click event: %d, %d, Mod: %d, Button: %d\n", x, y, ev.Modifiers(), ev.Buttons())
          }
      }
			}
		}
	}()

	return c.receive, c.send, nil
}

func (c *Clock) Render() []EventCell {
	rstring := time.Now().Format("2006-01-02 15:04:05")
	r := make([]EventCell, len(rstring))
	for i := range len(rstring){
    r[i] = EventCell{rune(rstring[i]), DEFAULT, "", c}
	}

	return r
}

func (c *Clock) Channels() (<-chan bool, chan<- Event) {
	return c.receive, c.send
}

func (c *Clock) Name() string {
	return "clock"
}
