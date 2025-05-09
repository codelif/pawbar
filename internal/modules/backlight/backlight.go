package backlight

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/jochenvg/go-udev"
)

var ICONS_BACKLIGHT = []rune{'󰃞', '󰃟', '󰃝', '󰃠'}

type Backlight struct {
	receive           chan bool
	send              chan modules.Event
	status            map[string]int
	backlight         string
	MaxBrightness     int
	currentBrightness int
	Type              string
	opts              Options
	initialOpts       Options
}

func New() modules.Module {
	return &Backlight{}
}

func (mod *Backlight) Dependencies() []string {
	return []string{}
}

func (mod *Backlight) Udev() (<-chan *udev.Device, error) {
	udevInstance := udev.Udev{}
	monitor := udevInstance.NewMonitorFromNetlink("udev")

	err := monitor.FilterAddMatchSubsystem("backlight")
	if err != nil {
		return nil, err
	}

	context_ := context.Background()
	devChan, errChan, err := monitor.DeviceChan(context_)
	if devChan == nil || errChan == nil {
		return nil, fmt.Errorf("failed to initialize backlight udev monitor")
	}

	inchan := make(chan *udev.Device)
	go func() {
		isRunning := true
		for isRunning {
			select {
			case d := <-devChan:
				if d == nil {
					isRunning = false
				} else {
					inchan <- d
				}
			case e := <-errChan:
				if e != nil {
					fmt.Println("udev monitor error:", e)
					isRunning = false
				}
			case <-context_.Done():
				isRunning = false
			}
		}
	}()

	return inchan, nil
}

func (mod *Backlight) getBacklight() (string, error) {
	basePath := "/sys/class/backlight/"
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return "", err
	}

	type backlightDevice struct {
		name          string
		devType       string
		maxBrightness int
	}
	var validDevices []backlightDevice

	for _, entry := range entries {
		devicePath := filepath.Join(basePath, entry.Name())
		typeData, err := os.ReadFile(filepath.Join(devicePath, "type"))
		if err != nil {
			continue
		}
		deviceType := strings.TrimSpace(string(typeData))

		maxData, err := os.ReadFile(filepath.Join(devicePath, "max_brightness"))
		if err != nil {
			continue
		}
		maxBrightness, err := strconv.Atoi(strings.TrimSpace(string(maxData)))
		if err != nil || maxBrightness == 0 {
			continue
		}

		validDevices = append(validDevices, backlightDevice{
			name:          entry.Name(),
			devType:       deviceType,
			maxBrightness: maxBrightness,
		})
	}

	if len(validDevices) == 0 {
		fmt.Println("No valid backlight devices found.")
		return "", fmt.Errorf("no valid backlight devices found")
	}

	selected := validDevices[0]
	for _, d := range validDevices {
		if d.devType == "raw" {
			selected = d
			break
		}
	}

	mod.Type = selected.devType
	mod.MaxBrightness = selected.maxBrightness
	return selected.name, nil
}

func (mod *Backlight) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *Backlight) Name() string {
	return "backlight"
}

func (mod *Backlight) Update() {
	if mod.backlight == "" {
		deviceName, err := mod.getBacklight()
		if err != nil {
			fmt.Println("Error getting backlight device:", err)
			return
		}
		mod.backlight = deviceName
	}

	if mod.status == nil {
		mod.status = make(map[string]int)
	}

	base := filepath.Join("/sys/class/backlight", mod.backlight)
	data, err := os.ReadFile(filepath.Join(base, "brightness"))
	if err != nil {
		fmt.Println("Error reading brightness:", err)
		return
	}
	now, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		fmt.Println("Error converting brightness:", err)
		return
	}
	mod.status["now"] = now
	mod.currentBrightness = now

	if mod.MaxBrightness == 0 {
		mdata, err := os.ReadFile(filepath.Join(base, "max_brightness"))
		if err == nil {
			maxVal, _ := strconv.Atoi(strings.TrimSpace(string(mdata)))
			mod.MaxBrightness = maxVal
		}
	}
	mod.status["max"] = mod.MaxBrightness
}

func (mod *Backlight) Render() []modules.EventCell {
	if mod.status == nil {
		return nil
	}
	now := mod.status["now"]
	maxVal := mod.status["max"]
	if maxVal == 0 {
		return nil
	}
	percent := (now * 100) / maxVal

	idx := (len(ICONS_BACKLIGHT) - 1) * percent / 100
	icon := ICONS_BACKLIGHT[idx]

	rch := vaxis.Characters(fmt.Sprintf("%c %d%%", icon, percent))
	r := make([]modules.EventCell, len(rch))
	for i, ch := range rch {
		r[i] = modules.EventCell{C: vaxis.Cell{Character: ch}, Mod: mod}
	}
	return r
}

func (mod *Backlight) Run() (<-chan bool, chan<- modules.Event, error) {
	mod.send = make(chan modules.Event)
	mod.receive = make(chan bool)

	mod.Update()

	uchan, err := mod.Udev()
	if err != nil {
		return nil, nil, err
	}

	go func() {
		for {
			select {
			case <-uchan:
				mod.Update()
				mod.receive <- true
			case <-mod.send:
			}
		}
	}()

	return mod.receive, mod.send, nil
}
