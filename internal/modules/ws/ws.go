package ws

import (
	"bytes"
	"fmt"
	"os"
	"strconv"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/config"
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

type Format int

func (f *Format) toggle() { *f ^= 1 }

const (
	FormatAll Format = iota
	FormatCurr
)

type backend interface {
	List() []Workspace
	Events() <-chan struct{}
	Goto(name string)
}

type Module struct {
	b       backend
	receive chan bool
	send    chan modules.Event
	bname   string
	format  Format

	opts        Options
	initialOpts Options
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
	mod.initialOpts = mod.opts

	go func() {
		render := mod.b.Events()
		for {
			select {
			case e := <-mod.send:
				switch ev := e.VaxisEvent.(type) {
				case vaxis.Mouse:
					if ev.EventType != vaxis.EventPress {
						break
					}
					btn := config.ButtonName(ev)

					if btn == "left" {
						go mod.b.Goto(e.Cell.Metadata)
					}
					if btn == "right" {
						mod.format.toggle()
						mod.receive <- true
					}

					if mod.opts.OnClick.Dispatch(btn, &mod.initialOpts, &mod.opts) {
						mod.receive <- true
					}

				case modules.FocusIn:
					if mod.opts.OnClick.HoverIn(&mod.opts) {
						mod.receive <- true
					}

				case modules.FocusOut:
					if mod.opts.OnClick.HoverOut(&mod.opts) {
						mod.receive <- true
					}
				}
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
		mod.bname = "hypr"
	} else if os.Getenv("I3SOCK") != "" || os.Getenv("SWAYSOCK") != "" {
		svc, ok := i3.Register()
		if !ok {
			return fmt.Errorf("Could not start i3 service.")
		}
		mod.b = newI3Backend(svc)
		mod.bname = "i3"
	} else {
		return fmt.Errorf("Could not find a wm backend for current environment.")
	}

	return nil
}

func (mod *Module) Render() []modules.EventCell {
	data := struct{ WSID string }{}
	format := mod.opts.Format

	var toRender []Workspace
	switch mod.format {
	case FormatAll:
		toRender = mod.b.List()

	case FormatCurr:
		for _, w := range mod.b.List() {
			if w.Active {
				toRender = []Workspace{w}
				break
			}
		}

	default:
		toRender = mod.b.List()
	}

	var cells []modules.EventCell
	for _, w := range toRender {
		wsName := w.Name
		if w.Special {
			wsName = "S"
		}
		var meta string
		if mod.bname == "hypr" {
			meta = strconv.Itoa(w.ID)
		} else {
			meta = wsName
		}
		style := vaxis.Style{
			Foreground: mod.opts.Fg.Go(),
			Background: mod.opts.Bg.Go(),
		}
		switch {
		case w.Special:
			style.Foreground = mod.opts.Special.Fg.Go()
			style.Background = mod.opts.Special.Bg.Go()
		case w.Active:
			style.Foreground = mod.opts.Active.Fg.Go()
			style.Background = mod.opts.Active.Bg.Go()
		case w.Urgent:
			style.Foreground = mod.opts.Urgent.Fg.Go()
			style.Background = mod.opts.Urgent.Bg.Go()
		}
		data.WSID = " " + wsName + " "
		var buf bytes.Buffer
		if err := format.Execute(&buf, data); err != nil {
			continue
		}

		// split into cells
		for _, ch := range vaxis.Characters(buf.String()) {
			cells = append(cells, modules.EventCell{
				C: vaxis.Cell{
					Character: ch,
					Style:     style,
				},
				Metadata:   meta,
				Mod:        mod,
				MouseShape: vaxis.MouseShapeClickable,
			})
		}
	}

	return cells
}
