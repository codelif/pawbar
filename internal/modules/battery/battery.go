package battery

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/jochenvg/go-udev"
)

var (
	ICONS_DISCHARGING = []rune{'󰂃', '󰁺', '󰁻', '󰁼', '󰁽', '󰁾', '󰁿', '󰂀', '󰂁', '󰂂', '󰁹'}
	ICONS_CHARGING    = []rune{'󰢟', '󰢜', '󰂆', '󰂇', '󰂈', '󰢝', '󰂉', '󰢞', '󰂊', '󰂋', '󰂅'}
)

func New() modules.Module {
	return &Battery{}
}

type Battery struct {
	receive chan bool
	send    chan modules.Event
	status  map[string]int
	battery string
	mains   string
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

	uchan, err := mod.Udev()
	if err != nil {
		return nil, nil, err
	}

	go func() {
		t := time.NewTicker(5000 * time.Millisecond)
		for {
			select {
			case <-t.C:
				if mod.Update() {
					mod.receive <- true
				}
			case <-uchan:
				if mod.Update() {
					mod.receive <- true
				}
			case <-mod.send:
			}
		}
	}()

	return mod.receive, mod.send, nil
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
	s := vaxis.Style{}
	icon := ' '
	if mod.status["mains"] == 1 {
		icon = ICONS_CHARGING[(len(ICONS_CHARGING)-1)*percent/100]
		if mod.status["charging"] == 0 {
			s.Foreground = modules.GOOD
		}
	} else {
		icon = ICONS_DISCHARGING[(len(ICONS_DISCHARGING)-1)*percent/100]
		if percent <= 15 {
			s.Foreground = modules.URGENT
		} else if percent <= 30 {
			s.Foreground = modules.WARNING
		}

	}

	rch := vaxis.Characters(fmt.Sprintf("%c %d%%", icon, percent))
	r := make([]modules.EventCell, len(rch))

	for i, ch := range rch {
		r[i] = modules.EventCell{C: vaxis.Cell{Character: ch, Style: s}, Mod: mod}
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
	}
	return battery, mains, nil
}
