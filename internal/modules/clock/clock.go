package clock

import (
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/itchyny/timefmt-go"
)

type ClockModule struct {
	receive chan bool
	send    chan modules.Event

	opts        Options
	initialOpts Options

	currentTickerInterval time.Duration
	ticker                *time.Ticker
}

func (mod *ClockModule) Dependencies() []string {
	return nil
}

func (mod *ClockModule) Run() (<-chan bool, chan<- modules.Event, error) {
	mod.receive = make(chan bool)
	mod.send = make(chan modules.Event)
	mod.initialOpts = mod.opts

	go func() {
		mod.currentTickerInterval = mod.opts.Tick.Go()
		mod.ticker = time.NewTicker(mod.currentTickerInterval)
		defer mod.ticker.Stop()
		for {
			select {
			case <-mod.ticker.C:
				mod.receive <- true
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
					mod.ensureTickInterval()

				case modules.FocusIn:
					if mod.opts.OnClick.HoverIn(&mod.opts) {
						mod.receive <- true
					}
					mod.ensureTickInterval()

				case modules.FocusOut:
					if mod.opts.OnClick.HoverOut(&mod.opts) {
						mod.receive <- true
					}
					mod.ensureTickInterval()
				}
			}
		}
	}()

	return mod.receive, mod.send, nil
}


func (mod *ClockModule) ensureTickInterval() {
	if mod.opts.Tick.Go() != mod.currentTickerInterval {
		mod.currentTickerInterval = mod.opts.Tick.Go()
		mod.ticker.Reset(mod.currentTickerInterval)
	}
}

func (mod *ClockModule) Render() []modules.EventCell {
	var s vaxis.Style
	s.Foreground = mod.opts.Fg.Go()
	s.Background = mod.opts.Bg.Go()

	rch := vaxis.Characters(timefmt.Format(time.Now(), mod.opts.Format))
	r := make([]modules.EventCell, len(rch))
	for i, ch := range rch {
		r[i] = modules.EventCell{
			C: vaxis.Cell{
				Character: ch,
				Style:     s,
			},
			Metadata:   "",
			Mod:        mod,
			MouseShape: mod.opts.Cursor.Go(),
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
