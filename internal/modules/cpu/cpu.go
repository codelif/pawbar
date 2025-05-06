package cpu

import (
	"bytes"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/mouse"
	"github.com/shirou/gopsutil/v3/cpu"
)

type CpuModule struct {
	receive chan bool
	send    chan modules.Event

	opts  Options
	cfg   Config
	mouse mouse.Mixin[Options]

	highStart     time.Time
	highTriggered bool
}

func (mod *CpuModule) Dependencies() []string {
	return nil
}

func (mod *CpuModule) Run() (<-chan bool, chan<- modules.Event, error) {
	mod.receive = make(chan bool)
	mod.send = make(chan modules.Event)

	mod.mouse = mouse.Mixin[Options]{
		Handlers: mod.cfg.OnClicks,
		Apply: func(ov *Options) bool {
			newOpts := mod.opts
			if !mouse.Overlay(&newOpts, ov) {
				return false
			}

			mod.cfg = newConfig(newOpts)
			mod.opts = newOpts
			return true
		},
    Cursor: mod.cfg.Cursor,
	}

	go func() {
		t := time.NewTicker(mod.cfg.Tick)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				mod.receive <- true
			case e := <-mod.send:
				switch ev := e.VaxisEvent.(type) {
				case vaxis.Mouse:

					if mod.mouse.Handle(ev) {
						mod.receive <- true
					}
				}
			}
		}
	}()

	return mod.receive, mod.send, nil
}

func (mod *CpuModule) Render() []modules.EventCell {
	percent, err := cpu.Percent(0, false)
	if err != nil || len(percent) == 0 {
		return nil
	}
	usage := int(percent[0])

	threshold := mod.cfg.ThrPercent
	if usage > threshold {
		if mod.highStart.IsZero() {
			mod.highStart = time.Now()
		} else if !mod.highTriggered && time.Since(mod.highStart) >= mod.cfg.ThrFor {
			mod.highTriggered = true
		}
	} else {
		mod.highStart = time.Time{}
		mod.highTriggered = false
	}

	style := vaxis.Style{}
	if mod.highTriggered {
		style.Foreground = mod.cfg.ThrColor
	} else {
		style.Foreground = mod.cfg.FgColor
	}

	style.Background = mod.cfg.BgColor

	var buf bytes.Buffer
	_ = mod.cfg.Tmpl.Execute(&buf, struct{ Percent int }{usage})

	rch := vaxis.Characters(buf.String())
	r := make([]modules.EventCell, len(rch))

	for i, ch := range rch {
		r[i] = modules.EventCell{C: vaxis.Cell{Character: ch, Style: style}, Mod: mod, MouseShape: mod.mouse.Cursor}
	}
	return r
}

func (mod *CpuModule) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *CpuModule) Name() string {
	return "cpu"
}
