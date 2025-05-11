package ram

import (
	"bytes"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/lookup/units"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/shirou/gopsutil/v3/mem"
)

type RamModule struct {
	receive chan bool
	send    chan modules.Event

	opts        Options
	initialOpts Options

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
	system := units.IEC
	if mod.opts.UseSI {
		system = units.SI
	}

	unit := mod.opts.Scale.Unit
	if mod.opts.Scale.Dynamic || mod.opts.Scale.Unit.Name == "" {
		unit = units.Choose(v.Total, system)
	}

	usedAbs := units.Format(v.Used, unit)
	freeAbs := units.Format(v.Available, unit)
	totalAbs := units.Format(v.Total, unit)

	usedPercent := int(v.UsedPercent)
	freePercent := 100 - usedPercent

	usage := usedPercent
	style := vaxis.Style{}
	if usage > mod.opts.Urgent.Percent.Go() {
		style.Foreground = mod.opts.Urgent.Fg.Go()
		style.Background = mod.opts.Urgent.Bg.Go()
	} else if usage > mod.opts.Warning.Percent.Go() {
		style.Foreground = mod.opts.Warning.Fg.Go()
		style.Background = mod.opts.Warning.Bg.Go()
	} else {
		style.Foreground = mod.opts.Fg.Go()
		style.Background = mod.opts.Bg.Go()
	}

	var buf bytes.Buffer

	err = mod.opts.Format.Execute(&buf, struct {
		Used, Free, Total        float64
		UsedPercent, FreePercent int
		Unit, Icon               string
	}{
		usedAbs, freeAbs, totalAbs,
		usedPercent, freePercent,
		unit.Name, mod.opts.Icon.Go(),
	})

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
