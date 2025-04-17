package i3title

import (
	"errors"
	"strings"

	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/services/hypr"
)

func New() modules.Module {
	return &i3Title{}
}

type i3Title struct {
	receive chan bool
	send    chan modules.Event
	hevent  chan hypr.HyprEvent
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
		return nil, nil, errors.New("Hypr service not available")
	}

	hyprtitle.receive = make(chan bool)
	hyprtitle.send = make(chan modules.Event)
	hyprtitle.hevent = make(chan i3.I3Event)
	activews := hypr.GetActiveWorkspace()
	clients := hypr.GetClients()

	hyprtitle.class = ""
	for _, c := range clients {
		if c.Address == activews.Lastwindow {
			hyprtitle.class = c.Class
		}
	}

	hyprtitle.title = hypr.GetActiveWorkspace().Lastwindowtitle
	service.RegisterChannel("activewindow", hyprtitle.hevent)

	go func() {
		for {
			select {
			case h := <-hyprtitle.hevent:
				hyprtitle.class, hyprtitle.title, _ = strings.Cut(h.Data, ",")
				hyprtitle.receive <- true
			case <-hyprtitle.send:
			}
		}
	}()

	return hyprtitle.receive, hyprtitle.send, nil
}
