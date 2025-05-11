package battery

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/utils"
	"github.com/jochenvg/go-udev"
)

func New() modules.Module {
	return &Battery{}
}

type Battery struct {
	receive  chan bool
	send     chan modules.Event
	status   map[string]int
	battery  string
	mains    string
	hoursRem int
	minsRem  int

	opts        Options
	initialOpts Options

	currentTickerInterval time.Duration
	ticker                *time.Ticker
}

func (mod *Battery) Dependencies() []string {
	return []string{}
}

func (mod *Battery) Run() (<-chan bool, chan<- modules.Event, error) {
	battery, mains, err := mod.GetSupplies()
	if err != nil {
		return nil, nil, err
	}

	mod.send = make(chan modules.Event)
	mod.receive = make(chan bool)
	mod.status = make(map[string]int)
	mod.status["now"] = 0
	mod.status["full"] = 100
	mod.status["mains"] = 0
	mod.status["charging"] = 0
	mod.battery = battery
	mod.mains = mains
	mod.Update()
	mod.initialOpts = mod.opts

	uchan, err := mod.Udev()
	if err != nil {
		return nil, nil, err
	}

	go func() {
		mod.currentTickerInterval = mod.opts.Tick.Go()
		mod.ticker = time.NewTicker(mod.currentTickerInterval)
		defer mod.ticker.Stop()
		for {
			select {
			case <-mod.ticker.C:
				if mod.Update() {
					mod.receive <- true
				}
			case <-uchan:
				if mod.Update() {
					mod.receive <- true
				}
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

func (mod *Battery) ensureTickInterval() {
	if mod.opts.Tick.Go() != mod.currentTickerInterval {
		mod.currentTickerInterval = mod.opts.Tick.Go()
		mod.ticker.Reset(mod.currentTickerInterval)
	}
}

func (mod *Battery) Update() bool {
	change := false
	file, err := os.Open(filepath.Join(mod.battery, "uevent"))
	if err != nil {
		return false
	}
	defer file.Close()

	var now, full, design, mains int
	var status string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "POWER_SUPPLY_ENERGY_NOW=") {
			now, _ = strconv.Atoi(strings.TrimPrefix(line, "POWER_SUPPLY_ENERGY_NOW="))
		} else if strings.HasPrefix(line, "POWER_SUPPLY_ENERGY_FULL=") {
			full, _ = strconv.Atoi(strings.TrimPrefix(line, "POWER_SUPPLY_ENERGY_FULL="))
		} else if strings.HasPrefix(line, "POWER_SUPPLY_CHARGE_NOW=") {
			now, _ = strconv.Atoi(strings.TrimPrefix(line, "POWER_SUPPLY_CHARGE_NOW="))
		} else if strings.HasPrefix(line, "POWER_SUPPLY_CHARGE_FULL=") {
			full, _ = strconv.Atoi(strings.TrimPrefix(line, "POWER_SUPPLY_CHARGE_FULL="))
		} else if strings.HasPrefix(line, "POWER_SUPPLY_ENERGY_FULL_DESIGN=") {
			design, _ = strconv.Atoi(strings.TrimPrefix(line, "POWER_SUPPLY_ENERGY_FULL_DESIGN="))
		} else if strings.HasPrefix(line, "POWER_SUPPLY_CHARGE_FULL_DESIGN=") {
			design, _ = strconv.Atoi(strings.TrimPrefix(line, "POWER_SUPPLY_CHARGE_FULL_DESIGN="))
		} else if strings.HasPrefix(line, "POWER_SUPPLY_STATUS=") {
			status = strings.TrimSpace(strings.TrimPrefix(line, "POWER_SUPPLY_STATUS="))
		}
	}

	percent_before := (mod.status["now"] * 100) / (mod.status["full"])
	percent_now := (now * 100) / (full)

	mod.status["now"] = now
	mod.status["full"] = full
	mod.status["design"] = design

	change = percent_before != percent_now

	m, _ := os.ReadFile(filepath.Join(mod.mains, "online"))
	mains, _ = strconv.Atoi(strings.TrimSpace(string(m)))
	if mod.status["mains"] != mains {
		mod.status["mains"] = mains
		change = true
	}

	charging := 0
	switch status {
	case "Charging":
		charging = 1
	case "Not charging", "Full", "Discharging":
		charging = 0
	}

	if mod.status["charging"] != charging {
		mod.status["charging"] = charging
		change = true
	}

	return change
}

func (mod *Battery) Render() []modules.EventCell {
	percent := (mod.status["now"] * 100) / (mod.status["full"])
	style := vaxis.Style{}
	icon := ' '
	if mod.status["mains"] == 1 {
		icons := mod.opts.IconsCharging
		icon = icons[utils.Clamp((len(icons)-1)*percent/100, 0, len(icons)-1)]
		if mod.status["charging"] == 0 || percent >= mod.opts.Optimal.Percent.Go() {
			style.Foreground = mod.opts.Optimal.Fg.Go()
		}
	} else {
		icons := mod.opts.IconsDischarging
		icon = icons[utils.Clamp((len(icons)-1)*percent/100, 0, len(icons)-1)]
		if percent <= mod.opts.Urgent.Percent.Go() {
			style.Foreground = mod.opts.Urgent.Fg.Go()
		} else if percent <= mod.opts.Warning.Percent.Go() {
			style.Foreground = mod.opts.Warning.Fg.Go()
		} else if percent >= mod.opts.Optimal.Percent.Go() {
			style.Foreground = mod.opts.Optimal.Fg.Go()
		}

	}

	var buf bytes.Buffer

	err := mod.opts.Format.Execute(&buf, struct {
		Icon        string
		UsedPercent int
		Hours       int
		Minutes     int
	}{
		Icon:        string(icon),
		UsedPercent: percent,
		Hours:       mod.hoursRem,
		Minutes:     mod.minsRem,
	})
	if err != nil {
		return nil
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

func (mod *Battery) Udev() (<-chan *udev.Device, error) {
	u := udev.Udev{}
	monitor := u.NewMonitorFromNetlink("udev")

	err := monitor.FilterAddMatchSubsystem("power_supply")
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	ch, errch, err := monitor.DeviceChan(ctx)
	if err != nil {
		return nil, err
	}

	uchan := make(chan *udev.Device)
	go func() {
		isRunning := true
		for isRunning {
			select {
			case d := <-ch:
				uchan <- d
			case err = <-errch:
				isRunning = false
			case <-ctx.Done():
				isRunning = false
			}
		}
	}()

	return uchan, nil
}

func (mod *Battery) GetSupplies() (string, string, error) {
	basePath := "/sys/class/power_supply/"
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return "", "", err
	}

	battery := ""
	mains := ""

	pref_bat_found := false

	for _, entry := range entries {
		typePath := filepath.Join(basePath, entry.Name(), "type")
		data, err := os.ReadFile(typePath)
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(data)) == "Battery" && !pref_bat_found {
			if strings.HasPrefix(entry.Name(), "BAT") {
				pref_bat_found = true
			}
			battery = filepath.Join(basePath, entry.Name())
		} else if strings.TrimSpace(string(data)) == "Mains" {
			mains = filepath.Join(basePath, entry.Name())
		}
	}

	if battery == "" || mains == "" {
		return "", "", fmt.Errorf("WARNING(battery): Battery or mains not available. Disabling.")
	} else {
		powerNowPath := filepath.Join(battery, "power_now")
		energyNowPath := filepath.Join(battery, "energy_now")

		powerNowData, err := os.ReadFile(powerNowPath)
		if err != nil {
			// fallback to current_now
			powerNowPath = filepath.Join(battery, "current_now")
			powerNowData, err = os.ReadFile(powerNowPath)
			if err != nil {
				utils.Logger.Println("cannot read power_now or current_now")
			}
		}

		energyNowData, err := os.ReadFile(energyNowPath)
		if err != nil {
			// fallback to charge_now
			energyNowPath = filepath.Join(battery, "charge_now")
			energyNowData, err = os.ReadFile(energyNowPath)
			if err != nil {
				utils.Logger.Println("cannot read energy_now or charge_now")
			}
		}

		powerNow, err := strconv.Atoi(strings.TrimSpace(string(powerNowData)))
		energyNow, err := strconv.Atoi(strings.TrimSpace(string(energyNowData)))
		if err != nil {
			utils.Logger.Println("conversion failed")
		}

		if powerNow == 0 {
			utils.Logger.Println("powerNow is zero, cannot compute time remaining")
		} else {

			timeRemainingHours := float64(energyNow) / float64(powerNow)

			mod.hoursRem = int(timeRemainingHours)
			mod.minsRem = int((timeRemainingHours - float64(mod.hoursRem)) * 60)
		}
	}
	return battery, mains, nil
}
