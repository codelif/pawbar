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

func (b *Battery) Dependencies() []string {
	return []string{}
}

func (b *Battery) Run() (<-chan bool, chan<- modules.Event, error) {
	battery, mains, err := b.GetSupplies()
	if err != nil {
		return nil, nil, err
	}

	b.send = make(chan modules.Event)
	b.receive = make(chan bool)
	b.status = make(map[string]int)
	b.status["now"] = 0
	b.status["full"] = 100
	b.status["mains"] = 0
	b.status["charging"] = 0
	b.battery = battery
	b.mains = mains
	b.Update()

	uchan, err := b.Udev()
	if err != nil {
		return nil, nil, err
	}

	go func() {
		t := time.NewTicker(5000 * time.Millisecond)
		for {
			select {
			case <-t.C:
				if b.Update() {
					b.receive <- true
				}
			case <-uchan:
				if b.Update() {
					b.receive <- true
				}
			case <-b.send:
			}
		}
	}()

	return b.receive, b.send, nil
}

func (b *Battery) Update() bool {
	change := false
	file, err := os.Open(filepath.Join(b.battery, "uevent"))
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

	percent_before := (b.status["now"] * 100) / (b.status["full"])
	percent_now := (now * 100) / (full)

	b.status["now"] = now
	b.status["full"] = full
	b.status["design"] = design

	change = percent_before != percent_now

	m, _ := os.ReadFile(filepath.Join(b.mains, "online"))
	mains, _ = strconv.Atoi(strings.TrimSpace(string(m)))
	if b.status["mains"] != mains {
		b.status["mains"] = mains
		change = true
	}

	charging := 0
	switch status {
	case "Charging":
		charging = 1
	case "Not charging", "Full", "Discharging":
		charging = 0
	}

	if b.status["charging"] != charging {
		b.status["charging"] = charging
		change = true
	}

	return change
}

func (b *Battery) Render() []modules.EventCell {
	percent := (b.status["now"] * 100) / (b.status["full"])
	s := modules.DEFAULT
	icon := ' '
	if b.status["mains"] == 1 {
		icon = ICONS_CHARGING[(len(ICONS_CHARGING)-1)*percent/100]
		if b.status["charging"] == 0 {
			s = modules.GOOD
		}
	} else {
		icon = ICONS_DISCHARGING[(len(ICONS_DISCHARGING)-1)*percent/100]
		if percent <= 15 {
			s = modules.URGENT
		} else if percent <= 30 {
			s = modules.WARNING
		}

	}

	rstring := fmt.Sprintf(" %d%%", percent)
	r := make([]modules.EventCell, len(rstring)+1)

	i := 0
	r[i] = modules.EventCell{C: icon, Style: s, Metadata: "", Mod: b}
	i++

	for _, ch := range rstring {
		r[i] = modules.EventCell{C: ch, Style: s, Metadata: "", Mod: b}
		i++
	}

	return r
}

func (b *Battery) Channels() (<-chan bool, chan<- modules.Event) {
	return b.receive, b.send
}

func (b *Battery) Name() string {
	return "battery"
}

func (b *Battery) Udev() (<-chan *udev.Device, error) {
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

func (b *Battery) GetSupplies() (string, string, error) {
	basePath := "/sys/class/power_supply/"
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return "", "", err
	}

	battery := ""
	mains := ""

	for _, entry := range entries {
		typePath := filepath.Join(basePath, entry.Name(), "type")
		data, err := os.ReadFile(typePath)
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(data)) == "Battery" {
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
