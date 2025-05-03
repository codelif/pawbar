package hyprtitle

import (
	"errors"
	"strings"

	"git.sr.ht/~rockorager/vaxis"
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

func (mod *HyprTitle) Dependencies() []string {
	return []string{"hypr"}
}

func (mod *HyprTitle) Run() (<-chan bool, chan<- modules.Event, error) {
	service, ok := hypr.Register()
	if !ok {
		return nil, nil, errors.New("Hypr service not available")
	}

	mod.receive = make(chan bool)
	mod.send = make(chan modules.Event)
	mod.hevent = make(chan hypr.HyprEvent)
	activews := hypr.GetActiveWorkspace()
	clients := hypr.GetClients()

	mod.class = ""
	for _, c := range clients {
		if c.Address == activews.Lastwindow {
			mod.class = c.Class
		}
	}

	mod.title = hypr.GetActiveWorkspace().Lastwindowtitle
	service.RegisterChannel("activewindow", mod.hevent)

	go func() {
		for {
			select {
			case h := <-mod.hevent:
				mod.class, mod.title, _ = strings.Cut(h.Data, ",")
				mod.receive <- true
			case <-mod.send:
			}
		}
	}()

	return mod.receive, mod.send, nil
}

func (mod *HyprTitle) Render() []modules.EventCell {
	var r []modules.EventCell

	styleBg := vaxis.Style{Foreground: modules.BLACK, Background: modules.COOL}

	if mod.class != "" {
		rch := vaxis.Characters(" " + mod.class + " ")
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

func (mod *HyprTitle) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *HyprTitle) Name() string {
	return "hyprtitle"
}
