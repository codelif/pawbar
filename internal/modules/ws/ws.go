package ws

import (
	"fmt"
	"os"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/services/hypr"
	"github.com/codelif/pawbar/internal/services/i3"
)

type Workspace struct {
	ID      int
	Name    string
	Active  bool
	Urgent  bool
	Special bool
}

type backend interface {
	List() []Workspace
	Events() <-chan struct{}
	Goto(name string)
}

type Module struct {
	b       backend
	receive chan bool
	send    chan modules.Event
}

func New() modules.Module { return &Module{} }

func (mod *Module) Name() string                                  { return "ws" }
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
			case e := <-mod.send:
				switch ev := e.VaxisEvent.(type) {
				case vaxis.Mouse:
					mod.handleMouseEvent(e, ev)
				}
			case <-render:
				mod.receive <- true
			}
		}
	}()

	return mod.receive, mod.send, nil
}

func (mod *Module) handleMouseEvent(e modules.Event, ev vaxis.Mouse) {
	if ev.EventType != vaxis.EventPress {
		return
	}

	switch ev.Button {
	case vaxis.MouseLeftButton:
		go mod.b.Goto(e.Cell.Metadata)
	}
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
	ws := mod.b.List()

	var r []modules.EventCell

	for _, w := range ws {
		wsName := w.Name
		if w.Special {
			wsName = "S"
		}

		style := vaxis.Style{}
		mouseShape := vaxis.MouseShapeClickable

		if w.Special {
			style = SPECIAL
			mouseShape = vaxis.MouseShapeDefault
		} else if w.Active {
			style = ACTIVE
			mouseShape = vaxis.MouseShapeDefault
		} else if w.Urgent {
			style = URGENT
		}

		for _, ch := range vaxis.Characters(" " + wsName + " ") {
			r = append(r, modules.EventCell{C: vaxis.Cell{Character: ch, Style: style}, Metadata: wsName, Mod: mod, MouseShape: mouseShape})
		}
	}

	return r
}

var (
	SPECIAL = vaxis.Style{Foreground: modules.ACTIVE, Background: modules.SPECIAL}
	ACTIVE  = vaxis.Style{Foreground: modules.BLACK, Background: modules.ACTIVE}
	URGENT  = vaxis.Style{Foreground: modules.BLACK, Background: modules.URGENT}
)
