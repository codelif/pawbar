package disk

import (
	"bytes"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/shirou/gopsutil/v3/disk"
)

type DiskModule struct {
	receive               chan bool
	send                  chan modules.Event
	format                Format
	opts                  Options
	initialOpts           Options
	currentTickerInterval time.Duration
	ticker                *time.Ticker
}

const (
	bitUnit = 1 << iota
	bitFree
)

type Format int

const (
	UsedPercent Format = iota
	UsedAbsolute
	FreePercent
	FreeAbsolute
)

func (f *Format) toggleUnit() { *f ^= bitUnit }
func (f *Format) toggleFree() { *f ^= bitFree }

const GiB = 1073741824.0

func New() modules.Module { return &DiskModule{} }

func (mod *DiskModule) Dependencies() []string { return nil }

func (mod *DiskModule) syncTemplate() {
	switch mod.format {
	case UsedPercent:
		mod.opts.Format = mod.initialOpts.Format

	case UsedAbsolute:
		if a := mod.opts.OnClick.Actions["left"]; a != nil &&
			len(a.Configs) > 0 && a.Configs[0].Format != nil {
			mod.opts.Format = *a.Configs[0].Format
		}

	case FreePercent:
		if a := mod.opts.OnClick.Actions["right"]; a != nil &&
			len(a.Configs) > 0 && a.Configs[0].Format != nil {
			mod.opts.Format = *a.Configs[0].Format
		}

	case FreeAbsolute:
		if a := mod.opts.OnClick.Actions["right"]; a != nil &&
			len(a.Configs) > 1 && a.Configs[1].Format != nil {
			mod.opts.Format = *a.Configs[1].Format
		}
	}
}

func (mod *DiskModule) Run() (<-chan bool, chan<- modules.Event, error) {
	mod.receive = make(chan bool)
	mod.send = make(chan modules.Event)
	mod.initialOpts = mod.opts
	mod.syncTemplate()

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

					changed := false
					switch btn {
					case "left":
						mod.format.toggleUnit()
						changed = true
					case "right":
						mod.format.toggleFree()
						changed = true
					case "middle":
						if mod.format != UsedPercent {
							mod.format = UsedPercent
							changed = true
						}
					}

					if mod.opts.OnClick.Dispatch(btn, &mod.initialOpts, &mod.opts) {
						changed = true
					}

					if changed {
						mod.syncTemplate() // keep template & mode in lock-step
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

func (mod *DiskModule) ensureTickInterval() {
	if d := mod.opts.Tick.Go(); d != mod.currentTickerInterval {
		mod.currentTickerInterval = d
		mod.ticker.Reset(d)
	}
}

func (mod *DiskModule) Render() []modules.EventCell {
	du, err := disk.Usage("/")
	if err != nil {
		return nil
	}

	freePercent := 100 - int(du.UsedPercent)
	freeAbsolute := float64(du.Free) / GiB
	usedPercent := int(du.UsedPercent)
	usedAbsolute := float64(du.Used) / GiB

	usage := int(du.UsedPercent)
	style := vaxis.Style{}
	if usage > 95 {
		style.Foreground = mod.opts.Threshold.FgUrg.Go()
		style.Background = mod.opts.Threshold.BgUrg.Go()
	} else if usage > 90 {
		style.Foreground = mod.opts.Threshold.FgWar.Go()
		style.Background = mod.opts.Threshold.BgWar.Go()
	} else {
		style.Foreground = mod.opts.Fg.Go()
		style.Background = mod.opts.Bg.Go()
	}

	var buf bytes.Buffer

	switch mod.format {
	case UsedPercent:
		_ = mod.opts.Format.Execute(&buf, struct{ Percent int }{usedPercent})

	case UsedAbsolute:
		if a := mod.opts.OnClick.Actions["left"]; a != nil &&
			len(a.Configs) > 0 && a.Configs[0].Format != nil {
			_ = a.Configs[0].Format.Execute(&buf,
				struct{ Absolute float64 }{usedAbsolute})
		}

	case FreePercent:
		if a := mod.opts.OnClick.Actions["right"]; a != nil &&
			len(a.Configs) > 0 && a.Configs[0].Format != nil {
			_ = a.Configs[0].Format.Execute(&buf,
				struct{ Percent int }{freePercent})
		}

	case FreeAbsolute:
		if a := mod.opts.OnClick.Actions["right"]; a != nil &&
			len(a.Configs) > 1 && a.Configs[1].Format != nil {
			_ = a.Configs[1].Format.Execute(&buf,
				struct{ Absolute float64 }{freeAbsolute})
		}
	}

	rch := vaxis.Characters(buf.String())
	out := make([]modules.EventCell, len(rch))
	for i, ch := range rch {
		out[i] = modules.EventCell{
			C:          vaxis.Cell{Character: ch, Style: style},
			Mod:        mod,
			MouseShape: mod.opts.Cursor.Go(),
		}
	}
	return out
}

func (mod *DiskModule) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *DiskModule) Name() string { return "disk" }

