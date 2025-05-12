package battery

import (
	"bytes"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/lookup/icons"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/utils"
)

func New() modules.Module {
	return &Battery{}
}

type Battery struct {
	receive chan bool
	send    chan modules.Event

	opts        Options
	initialOpts Options

	device UPowerDevice
}

func (mod *Battery) Dependencies() []string {
	return []string{}
}

func (mod *Battery) Run() (<-chan bool, chan<- modules.Event, error) {

	mod.send = make(chan modules.Event)
	mod.receive = make(chan bool)
	mod.initialOpts = mod.opts

	upower, uch, err := ConnectUPower()
	if err != nil {
		return nil, nil, err
	}

	mod.device, _ = GetDisplayDevice(upower)

	go func() {
		defer upower.Close()
		for {
			select {
			case sig := <-uch:
				HandleSignal(sig, &mod.device)
				mod.receive <- true
			case e := <-mod.send:
				switch ev := e.VaxisEvent.(type) {
				case vaxis.Mouse:
					if ev.EventType != vaxis.EventPress {
						continue
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
			}
		}
	}()

	return mod.receive, mod.send, nil
}
func pickThreshold(p int, th []ThresholdOptions) *ThresholdOptions {
	for _, t := range th {
		matchUp := t.Direction.IsUp() && p >= t.Percent.Go()
		matchDown := !t.Direction.IsUp() && p <= t.Percent.Go()
		if matchUp || matchDown {
			return &t
		}
	}
	return nil
}
func (mod *Battery) Render() []modules.EventCell {
	percent := int(mod.device.Percentage)
	style := vaxis.Style{}

	icon := ' '
	eta := 0

	if mod.device.State == StateCharging || mod.device.State == StateFullyCharged {
		icon = icons.Choose(mod.opts.Charging.Icons, percent)
		eta = int(mod.device.TimeToFull)
		style.Foreground = mod.opts.Charging.Fg.Go()
		style.Background = mod.opts.Charging.Bg.Go()
	}

	if mod.device.State == StateDischarging {
		icon = icons.Choose(mod.opts.Discharging.Icons, percent)
		eta = int(mod.device.TimeToEmpty)
		style.Foreground = mod.opts.Discharging.Fg.Go()
		style.Background = mod.opts.Discharging.Bg.Go()
	}

	t := pickThreshold(percent, mod.opts.Thresholds)
	if t != nil {
		style.Foreground = t.Fg.Go()
		style.Background = t.Bg.Go()
	}

	if mod.device.State == StateFullyCharged {
		style.Foreground = mod.opts.Charged.Fg.Go()
		style.Background = mod.opts.Charged.Bg.Go()
		icon = mod.opts.Charged.Icon
	}
  
  // TODO: make config items implement IsZeroer
  //       to save my soul
	if mod.opts.Fg.Go() != vaxis.Color(0) {
		style.Foreground = mod.opts.Fg.Go()
	}
	if mod.opts.Bg.Go() != vaxis.Color(0) {
		style.Background = mod.opts.Bg.Go()
	}

	var buf bytes.Buffer

	err := mod.opts.Format.Execute(&buf, struct {
		Icon    string
		Percent int
		Hours   int
		Minutes int
	}{
		Icon:    string(icon),
		Percent: percent,
		Hours:   eta / 3600,
		Minutes: (eta / 60) % 60,
	})
	if err != nil {
		utils.Logger.Printf("battery: render error: %v", err)
	}

	rch := vaxis.Characters(buf.String())
	r := make([]modules.EventCell, len(rch))

	for i, ch := range rch {
		r[i] = modules.EventCell{C: vaxis.Cell{Character: ch, Style: style}, Mod: mod, MouseShape: mod.opts.Cursor.Go()}
	}
	return r
}

func (mod *Battery) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *Battery) Name() string {
	return "battery"
}
