package main

import "strings"

func NewHyprTitle() Module {
	return &HyprTitle{hypr: hypr}
}

type HyprTitle struct {
	hypr    *Hypr
	receive chan bool
	send    chan Event
	hevent  chan HyprEvent
	title   string
}

func (hyprtitle *HyprTitle) Run() (<-chan bool, chan<- Event, error) {
	hyprtitle.receive = make(chan bool)
	hyprtitle.send = make(chan Event)
	hyprtitle.hevent = make(chan HyprEvent)
	hyprtitle.title = GetActiveWorkspace().Lastwindowtitle
	hyprtitle.hypr.RegisterChannel("activewindow", hyprtitle.hevent)

	go func() {
		for {
			select {
			case h := <-hyprtitle.hevent:
				_, title, _ := strings.Cut(h.data, ",")
				hyprtitle.title = title
				hyprtitle.receive <- true
			case <-hyprtitle.send:
			}
		}
	}()

	return hyprtitle.receive, hyprtitle.send, nil
}

func (hyprtitle *HyprTitle) Render() []EventCell {
	rstring := " " + hyprtitle.title
	r := make([]EventCell, len(rstring))
	for i := range len(rstring) {
		r[i] = EventCell{rune(rstring[i]), DEFAULT, "", hyprtitle}
	}

	return r
}

func (hyprtitle *HyprTitle) Channels() (<-chan bool, chan<- Event) {
	return hyprtitle.receive, hyprtitle.send
}

func (hyprtitle *HyprTitle) Name() string {
	return "hypr-title"
}
