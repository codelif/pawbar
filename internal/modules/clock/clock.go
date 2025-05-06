package clock

import (
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/itchyny/timefmt-go"
)

type ClockModule struct {
	receive chan bool
	send    chan modules.Event

	cfg        config
	leftIdx    int
	rightIdx   int
	usingRight bool
}

func (mod *ClockModule) Dependencies() []string {
	return nil
}

func (mod *ClockModule) Run() (<-chan bool, chan<- modules.Event, error) {
	mod.receive = make(chan bool)
	mod.send = make(chan modules.Event)

	go func() {
		t := time.NewTicker(mod.cfg.tick)
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

// this is a blocking function, only use it in event loop
func (mod *ClockModule) handleMouseEvent(ev vaxis.Mouse) {
	if ev.EventType == vaxis.EventPress {
		switch ev.Button {
		case vaxis.MouseLeftButton:
			mod.usingRight = false
			mod.leftIdx = (mod.leftIdx + 1) % len(mod.cfg.left)
		case vaxis.MouseRightButton:
			if len(mod.cfg.right) == 0 {
				break
			}
			if !mod.usingRight {
				mod.rightIdx = 0
			} else {
				mod.rightIdx = (mod.rightIdx + 1) % len(mod.cfg.right)
			}
			mod.usingRight = true
		case vaxis.MouseMiddleButton:
			mod.usingRight = false
			mod.leftIdx = 0
		}

		mod.receive <- true
	}
}

func (mod *ClockModule) layout() string {
	if mod.usingRight && len(mod.cfg.right) > 0 {
		return mod.cfg.right[mod.rightIdx]
	}
	return mod.cfg.left[mod.leftIdx]
}

func (mod *ClockModule) Render() []modules.EventCell {
	rch := vaxis.Characters(timefmt.Format(time.Now(), mod.layout()))
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
