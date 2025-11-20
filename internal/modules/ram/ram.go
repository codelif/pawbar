package ram

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/nekorg/pawbar/internal/config"
	"github.com/nekorg/pawbar/internal/lookup/units"
	"github.com/nekorg/pawbar/internal/modules"
)

type virtualMemoryStat struct {
	Total       uint64
	Available   uint64
	Used        uint64
	UsedPercent float64
}

func virtualMemory() (*virtualMemoryStat, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var memTotal, memAvailable uint64

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				return nil, errors.New("malformed MemTotal line in /proc/meminfo")
			}
			v, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return nil, err
			}
			memTotal = v * 1024
		}
		if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				return nil, errors.New("malformed MemAvailable line in /proc/meminfo")
			}
			v, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return nil, err
			}
			memAvailable = v * 1024
		}
		if memTotal != 0 && memAvailable != 0 {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if memTotal == 0 {
		return nil, errors.New("MemTotal not found in /proc/meminfo")
	}

	used := memTotal - memAvailable
	usedPercent := float64(used) * 100.0 / float64(memTotal)

	return &virtualMemoryStat{
		Total:       memTotal,
		Available:   memAvailable,
		Used:        used,
		UsedPercent: usedPercent,
	}, nil
}

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

func (mod *RamModule) Render() []modules.EventCell {
	v, err := virtualMemory()
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

	t := pickThreshold(usage, mod.opts.Thresholds)

	if t != nil {
		style.Foreground = t.Fg.Go()
		style.Background = t.Bg.Go()
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
