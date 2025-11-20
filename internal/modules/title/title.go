package title

import (
	"bytes"
	"fmt"
	"os"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/nekorg/pawbar/internal/config"
	"github.com/nekorg/pawbar/internal/modules"
	"github.com/nekorg/pawbar/internal/services/hypr"
	"github.com/nekorg/pawbar/internal/services/i3"
)

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

	opts        Options
	initialOpts Options
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
	win := mod.b.Window()
	var cells []modules.EventCell

	if win.Class != "" {
		style := vaxis.Style{
			Foreground: mod.opts.Class.Fg.Go(),
			Background: mod.opts.Class.Bg.Go(),
		}

		var buf bytes.Buffer
		_ = mod.opts.Class.Format.Execute(&buf, struct{ Class string }{
			Class: " " + win.Class + " ",
		})

		for _, ch := range vaxis.Characters(buf.String()) {
			cells = append(cells, modules.EventCell{
				C: vaxis.Cell{
					Character: ch,
					Style:     style,
				},
				Mod:        mod,
				MouseShape: vaxis.MouseShapeDefault,
			})
		}
	}

	if win.Title != "" && win.Class != "" {
		style := vaxis.Style{
			Foreground: mod.opts.Title.Fg.Go(),
			Background: mod.opts.Title.Bg.Go(),
		}

		var buf bytes.Buffer
		_ = mod.opts.Title.Format.Execute(&buf, struct{ Title string }{
			Title: win.Title,
		})
		cells = append(cells, modules.EventCell{C: vaxis.Cell{Character: vaxis.Character{Grapheme: " ", Width: 1}}, Mod: mod})
		for _, ch := range vaxis.Characters(buf.String()) {
			cells = append(cells, modules.EventCell{
				C: vaxis.Cell{
					Character: ch,
					Style:     style,
				},
				Mod:        mod,
				MouseShape: vaxis.MouseShapeDefault,
			})
		}
	}

	return cells
}
