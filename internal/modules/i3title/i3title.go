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
	receive  chan bool
	send     chan modules.Event
	ievent   chan interface{}
	ievent2  chan interface{}
	instance string
	title    string
}

func (mod *i3Title) Dependencies() []string {
	return []string{"i3"}
}

func (mod *i3Title) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *i3Title) Name() string {
	return "i3title"
}

func (mod *i3Title) Run() (<-chan bool, chan<- modules.Event, error) {
	service, ok := i3.GetService()
	if !ok {
		return nil, nil, errors.New("i3 service not available")
	}

	mod.receive = make(chan bool)
	mod.send = make(chan modules.Event)
	mod.ievent = make(chan interface{})
	mod.ievent2 = make(chan interface{})

	mod.instance, mod.title = i3.GetTitleClass()

	service.RegisterChannel("activeWindow", mod.ievent)
	service.RegisterChannel("workspaces", mod.ievent2)

	go func() {
		for {
			select {
			case ev := <-mod.ievent:
				switch evt := ev.(type) {
				case i3.I3WEvent:
					// Handle window event
					mod.instance, mod.title = i3.GetTitleClass()
					mod.receive <- true
				default:
					fmt.Println("Received unknown type on window event channel:", evt)
				}

			case ev := <-mod.ievent2:
				switch evt := ev.(type) {
				case i3.I3Event:
					// Handle workspace event
					mod.instance, mod.title = i3.GetTitleClass()
					mod.receive <- true
				default:
					fmt.Println("Received unknown type on workspace event channel:", evt)
				}

			case <-mod.send:
			}
		}
	}()

	return mod.receive, mod.send, nil
}

func (mod *i3Title) Render() []modules.EventCell {
	var r []modules.EventCell
	styleBg := vaxis.Style{Foreground: modules.BLACK, Background: modules.COOL}

	if mod.instance != "" {
		rch := vaxis.Characters(" " + mod.instance + " ")
		for _, ch := range rch {
			r = append(r, modules.EventCell{C: vaxis.Cell{Character: ch, Style: styleBg}, Mod: mod})
		}
		r = append(r, modules.EventCell{C: vaxis.Cell{Character: vaxis.Character{Grapheme: " ", Width: 1}}, Mod: mod})

	}
	for _, ch := range vaxis.Characters(mod.title) {
		r = append(r, modules.EventCell{C: vaxis.Cell{Character: ch}, Mod: mod})
	}

	return r
}
