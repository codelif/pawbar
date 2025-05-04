package title

import (
	"fmt"
	"os"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/services/hypr"
	"github.com/codelif/pawbar/internal/services/i3"
	"gopkg.in/yaml.v3"
)

func init() {
	config.Register("title", func(n *yaml.Node) (modules.Module, error) {
		return &Module{}, nil
	})
}

type Window struct {
	Title string
	Class string
}

type backend interface {
	Window() Window
	Events() <-chan struct{}
}

type Module struct {
	b       backend
	receive chan bool
	send    chan modules.Event
}

func New() modules.Module { return &Module{} }

func (mod *Module) Name() string                                  { return "title" }
func (mod *Module) Dependencies() []string                        { return nil }
func (mod *Module) Channels() (<-chan bool, chan<- modules.Event) { return mod.receive, mod.send }

func (mod *Module) Run() (<-chan bool, chan<- modules.Event, error) {
	err := mod.selectBackend()
	if err != nil {
		return nil, nil, err
	}

	mod.receive = make(chan bool)
	mod.send = make(chan modules.Event)

	go func() {
		render := mod.b.Events()
		for {
			select {
			case <-mod.send:
			case <-render:
				mod.receive <- true
			}
		}
	}()

	return mod.receive, mod.send, nil
}

func (mod *Module) selectBackend() error {
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		svc, ok := hypr.Register()
		if !ok {
			return fmt.Errorf("Could not start hypr service.")
		}
		mod.b = newHyprBackend(svc)
	} else if os.Getenv("I3SOCK") != "" || os.Getenv("SWAYSOCK") != "" {
		svc, ok := i3.Register()
		if !ok {
			return fmt.Errorf("Could not start i3 service.")
		}
		mod.b = newI3Backend(svc)
	} else {
		return fmt.Errorf("Could not find a wm backend for current environment.")
	}

	return nil
}

func (mod *Module) Render() []modules.EventCell {
	var r []modules.EventCell
	win := mod.b.Window()
	styleBg := vaxis.Style{Foreground: modules.BLACK, Background: modules.COOL}

	if win.Class != "" {
		rch := vaxis.Characters(" " + win.Class + " ")
		for _, ch := range rch {
			r = append(r, modules.EventCell{C: vaxis.Cell{Character: ch, Style: styleBg}, Mod: mod})
		}
		r = append(r, modules.EventCell{C: vaxis.Cell{Character: vaxis.Character{Grapheme: " ", Width: 1}}, Mod: mod})

		for _, ch := range vaxis.Characters(win.Title) {
			r = append(r, modules.EventCell{C: vaxis.Cell{Character: ch}, Mod: mod})
		}
	}

	return r
}
