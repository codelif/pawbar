package i3title

import (
	"errors"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/services/i3"
)

func New() modules.Module {
	return &i3Title{}
}

type i3Title struct {
	receive chan bool
	send    chan modules.Event
	ievent  chan i3.I3WEvent
	ievent2 chan i3.I3Event
	class   string
	title   string
}

func (it *i3Title) Dependencies() []string {
	return []string{"i3"}
}

func (it *i3Title) Channels() (<-chan bool, chan<- modules.Event) {
	return it.receive, it.send
}

func (it *i3Title) Name() string {
	return "i3title"
}

func (it *i3Title) Run() (<-chan bool, chan<- modules.Event, error) {
	service, ok := i3.GetService()
	if !ok {
		return nil, nil, errors.New("i3 service not available")
	}

	it.receive = make(chan bool)
	it.send = make(chan modules.Event)
	it.ievent = make(chan i3.I3WEvent)
	it.ievent2 = make(chan i3.I3Event)

	it.class, it.title= i3.GetTitleClass() 

	service.RegisterWChannel("activeWindow", it.ievent)
	service.RegisterChannel("workspaces", it.ievent2)

	go func() {
		for {
			select {
			case <- it.ievent:
				it.class, it.title= i3.GetTitleClass()
				it.receive <- true
			case <- it.ievent2:
				it.class, it.title= i3.GetTitleClass()
				it.receive <- true
			case <-it.send:
			}
		}
	}()

	return it.receive, it.send, nil
}

func (it *i3Title) Render() []modules.EventCell {
	var r []modules.EventCell
	if it.class != "" {
		r = append(r, modules.EventCell{C: ' ', Style: modules.COOL.Reverse(true), Metadata: "", Mod: it})
		for _, ch := range it.class {
			r = append(r, modules.EventCell{C: ch, Style: modules.COOL.Reverse(true), Metadata: "", Mod: it})
		}
		r = append(r, modules.EventCell{C: ' ', Style: modules.COOL.Reverse(true), Metadata: "", Mod: it})
		r = append(r, modules.EventCell{C: ' ', Style: modules.DEFAULT, Metadata: "", Mod: it})
	}
	for _, ch := range it.title {
		r = append(r, modules.EventCell{C: ch, Style: modules.DEFAULT, Metadata: "", Mod: it})
	}

	return r
}



