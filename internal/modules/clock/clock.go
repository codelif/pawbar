package clock

import (
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
	"gopkg.in/yaml.v3"
)

func init() {
	config.Register("clock", func(n *yaml.Node) (modules.Module, error) {
		return &ClockModule{}, nil
	})
}

func New() modules.Module {
	return &ClockModule{}
}

type Format int

const (
	FormatDefault Format = iota
	FormatAlt1
	FormatAlt2
)

type ClockModule struct {
	receive chan bool
	send    chan modules.Event
	format  Format
}

func (mod *ClockModule) Dependencies() []string {
	return nil
}

func (mod *ClockModule) Run() (<-chan bool, chan<- modules.Event, error) {
	mod.receive = make(chan bool)
	mod.send = make(chan modules.Event)

	go func() {
		t := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-t.C:
				mod.receive <- true
			case e := <-mod.send:
				switch ev := e.VaxisEvent.(type) {
				case vaxis.Mouse:
					mod.handleMouseEvent(ev)
				}
			}
		}
	}()

	return mod.receive, mod.send, nil
}

func (mod *ClockModule) timeFormatString() string {
	switch mod.format {
	case FormatDefault:
		return time.Now().Format("2006-01-02 15:04:05")
	case FormatAlt1:
		return time.Now().Format("Mon 15:04")
	case FormatAlt2:
		return time.Now().Format("2 January 2006 Monday 15:04")

	}
	return time.Now().Format("2006-01-02 15:04:05")
}

// this is a blocking function, only use it in event loop
func (mod *ClockModule) handleMouseEvent(ev vaxis.Mouse) {
	if ev.EventType == vaxis.EventPress {
		switch ev.Button {
		case vaxis.MouseLeftButton:
			mod.cycle()
			mod.receive <- true
		}
	}
}

func (mod *ClockModule) cycle() {
	switch mod.format {
	case FormatDefault:
		mod.format = FormatAlt1
	case FormatAlt1:
		mod.format = FormatAlt2
	case FormatAlt2:
		mod.format = FormatDefault
	}
}

func (mod *ClockModule) Render() []modules.EventCell {
	rch := vaxis.Characters(mod.timeFormatString())
	r := make([]modules.EventCell, len(rch))
	for i, ch := range rch {
		r[i] = modules.EventCell{
			C: vaxis.Cell{
				Character: ch,
			},
			Metadata:   "",
			Mod:        mod,
			MouseShape: vaxis.MouseShapeClickable,
		}
	}
	return r
}

func (mod *ClockModule) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *ClockModule) Name() string {
	return "clock"
}
