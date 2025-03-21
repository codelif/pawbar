package main

import (
	"time"
)

func NewClock() Module {
	return &Clock{}
}

type Clock struct {
	receive chan bool
	send    chan Event
}

func (c *Clock) Run() (<-chan bool, chan<- Event,  error) {
	c.receive = make(chan bool)
	c.send = make(chan Event)

	go func() {
		t := time.NewTicker(time.Second)
		for {
			select {
			case <-t.C:
				c.receive <- true
			case <-c.send:
			}
		}
	}()

	return c.receive, c.send, nil
}

func (c *Clock) Render() []EventCell {
	rstring := time.Now().Format("2006-01-02 15:04:05")
	r := make([]EventCell, len(rstring))
	for _, char := range time.Now().Format("2006-01-02 15:04:05") {
		r = append(r, EventCell{char, DEFAULT, c, ""})
	}

	return r
}

func (c *Clock) Channels() (<-chan bool, chan<- Event){
  return c.receive, c.send
}

// Returns printable name
func (c *Clock) Name() string {
  return "clock"
}
