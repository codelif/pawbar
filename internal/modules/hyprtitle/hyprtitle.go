package hyprtitle

import (
	"errors"
	"strings"

	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/services/hypr"
)

func New() modules.Module {
	return &HyprTitle{}
}

type HyprTitle struct {
	receive chan bool
	send    chan modules.Event
	hevent  chan hypr.HyprEvent
	class   string
	title   string
}

func (ht *HyprTitle) Dependencies() []string {
	return []string{"hypr"}
}
func (hyprtitle *HyprTitle) Run() (<-chan bool, chan<- modules.Event, error) {
	service, ok := hypr.GetService()
	if !ok {
		return nil, nil, errors.New("Hypr service not available")
	}

	hyprtitle.receive = make(chan bool)
	hyprtitle.send = make(chan modules.Event)
	hyprtitle.hevent = make(chan hypr.HyprEvent)
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

func (hyprtitle *HyprTitle) Render() []modules.EventCell {
	var r []modules.EventCell
	if hyprtitle.class != "" {
		r = append(r, modules.EventCell{C: ' ', Style: modules.COOL.Reverse(true), Metadata: "", Mod: hyprtitle})
		for _, ch := range hyprtitle.class {
			r = append(r, modules.EventCell{C: ch, Style: modules.COOL.Reverse(true), Metadata: "", Mod: hyprtitle})
		}
		r = append(r, modules.EventCell{C: ' ', Style: modules.COOL.Reverse(true), Metadata: "", Mod: hyprtitle})
		r = append(r, modules.EventCell{C: ' ', Style: modules.DEFAULT, Metadata: "", Mod: hyprtitle})
	}
	for _, ch := range hyprtitle.title {
		r = append(r, modules.EventCell{C: ch, Style: modules.DEFAULT, Metadata: "", Mod: hyprtitle})
	}

	return r
}

func (hyprtitle *HyprTitle) Channels() (<-chan bool, chan<- modules.Event) {
	return hyprtitle.receive, hyprtitle.send
}

func (hyprtitle *HyprTitle) Name() string {
	return "hyprtitle"
}
