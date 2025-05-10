package ram

import (
	"bytes"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/shirou/gopsutil/v3/mem"
)

type Format int

const (
	FormatPercentage Format = iota
	FormatAbsolute
)

func (f *Format) toggle() { *f ^= 1 }

func New() modules.Module {
	return &RamModule{}
}

type RamModule struct {
	receive               chan bool
	send                  chan modules.Event
	format                Format
	opts                  Options
	initialOpts           Options
	currentTickerInterval time.Duration
	ticker                *time.Ticker
}

func (mod *RamModule) Dependencies() []string {
	return nil
}

func (mod *RamModule) Run() (<-chan bool, chan<- modules.Event, error) {
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
					if btn == "left" {
						mod.format.toggle()
					}
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

func (mod *RamModule) ensureTickInterval() {
	if mod.opts.Tick.Go() != mod.currentTickerInterval {
		mod.currentTickerInterval = mod.opts.Tick.Go()
		mod.ticker.Reset(mod.currentTickerInterval)
	}
}

func (mod *RamModule) Render() []modules.EventCell {
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil
	}

	valuePercent := int(v.UsedPercent)
	valueAbsolute := float64(v.Used) / 1073741824.00

	style := vaxis.Style{}
	if valuePercent > 90 {
		style.Foreground = mod.opts.Threshold.FgUrg.Go()
		style.Background = mod.opts.Threshold.BgUrg.Go()
	} else if valuePercent > 80 {
		style.Foreground = mod.opts.Threshold.FgWar.Go()
		style.Background = mod.opts.Threshold.BgWar.Go()
	} else {
		style.Foreground = mod.opts.Fg.Go()
		style.Background = mod.opts.Bg.Go()

	}

	var buf bytes.Buffer
	switch mod.format {
	case FormatPercentage:
		_ = mod.opts.Format.Execute(&buf, struct{ Percent int }{valuePercent})
	case FormatAbsolute:
		leftAction := mod.opts.OnClick.Actions["left"]
		if leftAction != nil && len(leftAction.Configs) > 0 {
			_ = leftAction.Configs[0].Format.Execute(&buf, struct{ Absolute float64 }{valueAbsolute})
		}
	}

	rch := vaxis.Characters(buf.String())
	r := make([]modules.EventCell, len(rch))

	for i, ch := range rch {
		r[i] = modules.EventCell{C: vaxis.Cell{Character: ch, Style: style}, Mod: mod, MouseShape: mod.opts.Cursor.Go()}
	}
	return r
}

func (mod *RamModule) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *RamModule) Name() string {
	return "ram"
}
