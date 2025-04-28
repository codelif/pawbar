package i3title

import (
	"errors"
	"fmt"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/services/i3"
)

func New() modules.Module {
	return &i3Title{}
}

type i3Title struct {
	receive    chan bool
	send       chan modules.Event
	ievent     chan interface{}
	ievent2    chan interface{}
	instance   string
	title      string
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
	it.ievent = make(chan interface{})
	it.ievent2 = make(chan interface{})

	it.instance, it.title= i3.GetTitleClass() 

	service.RegisterChannel("activeWindow", it.ievent)
	service.RegisterChannel("workspaces", it.ievent2)

	go func() {
	for {
		select {
		case ev := <-it.ievent:
			switch evt := ev.(type) {
			case i3.I3WEvent:
				// Handle window event
				it.instance, it.title = i3.GetTitleClass()
				it.receive <- true
			default:
				fmt.Println("Received unknown type on window event channel:", evt)
			}

		case ev := <-it.ievent2:
			switch evt := ev.(type) {
			case i3.I3Event:
				// Handle workspace event
				it.instance, it.title = i3.GetTitleClass()
				it.receive <- true
			default:
				fmt.Println("Received unknown type on workspace event channel:", evt)
			}

		case <-it.send:	
		}
	}
}()

	return it.receive, it.send, nil
}

func (it *i3Title) Render() []modules.EventCell {
	var r []modules.EventCell
  styleBg := vaxis.Style{Foreground: modules.BLACK, Background: modules.COOL}

	if it.instance != "" {
		r = append(r, modules.EventCell{C: vaxis.Cell{Character: vaxis.Character{Grapheme: " ", Width: 1}, Style: styleBg}, Mod: it})
		for _, ch := range it.instance {
      r = append(r, modules.EventCell{C: vaxis.Cell{Character: vaxis.Character{Grapheme: string(ch), Width: 1}, Style: styleBg}, Mod: it})
		}
		r = append(r, modules.EventCell{C: vaxis.Cell{Character: vaxis.Character{Grapheme: " ", Width: 1}, Style: styleBg}, Mod: it})
		r = append(r, modules.EventCell{C: vaxis.Cell{Character: vaxis.Character{Grapheme: " ", Width: 1}}, Mod: it})
	}
	for _, ch := range it.title {
    r = append(r, modules.EventCell{C: vaxis.Cell{Character: vaxis.Character{Grapheme: string(ch), Width: 1}}, Mod: it})
	}

	return r
}



