package backlight

import (
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

var ICONS_BACKLIGHT = []rune{'🌑', '🌒', '🌓', '🌔', '🌕'}

type Backlight struct {
	receive           chan bool
	send              chan modules.Event
	status            map[string]int
	backlight         string
	MaxBrightness     int
	currentBrightness int
	Type              string
}

func New() modules.Module {
	return &Backlight{}
}

func (b *Backlight) Dependencies() []string {
	return []string{}
}

func (b *Backlight) Udev() (<-chan *udev.Device, error) {
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

func (b *Backlight) getBacklight() (string, error) {
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

	b.Type = selected.devType
	b.MaxBrightness = selected.maxBrightness
	return selected.name, nil
}

func (b *Backlight) Channels() (<-chan bool, chan<- modules.Event) {
	return b.receive, b.send
}

func (b *Backlight) Name() string {
	return "backlight"
}

func (b *Backlight) Update() {
	if b.backlight == "" {
		deviceName, err := b.getBacklight()
		if err != nil {
			fmt.Println("Error getting backlight device:", err)
			return
		}
		b.backlight = deviceName
	}

	if b.status == nil {
		b.status = make(map[string]int)
	}

	base := filepath.Join("/sys/class/backlight", b.backlight)
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
	b.status["now"] = now
	b.currentBrightness = now

	if b.MaxBrightness == 0 {
		mdata, err := os.ReadFile(filepath.Join(base, "max_brightness"))
		if err == nil {
			maxVal, _ := strconv.Atoi(strings.TrimSpace(string(mdata)))
			b.MaxBrightness = maxVal
		}
	}
	b.status["max"] = b.MaxBrightness
}

func (b *Backlight) Render() []modules.EventCell {
	if b.status == nil {
		return nil
	}
	now := b.status["now"]
	maxVal := b.status["max"]
	if maxVal == 0 {
		return nil
	}
	percent := (now * 100) / maxVal

	idx := (len(ICONS_BACKLIGHT) - 1) * percent / 100
	icon := ICONS_BACKLIGHT[idx]
	rstring := fmt.Sprintf(" %d%%", percent)

	r := make([]modules.EventCell, len(rstring)+1)
	i := 0
	r[i] = modules.EventCell{C: icon, Style: modules.DEFAULT, Metadata: "", Mod: b}
	i++

	for _, ch := range rstring {
		r[i] = modules.EventCell{C: ch, Style: modules.DEFAULT, Metadata: "", Mod: b}
		i++
	}
	return r
}

func (b *Backlight) Run() (<-chan bool, chan<- modules.Event, error) {
	b.send = make(chan modules.Event)
	b.receive = make(chan bool)

	b.Update()

	uchan, err := b.Udev()
	if err != nil {
		return nil, nil, err
	}

	go func() {
		for {
			select {
			case <-uchan:
				b.Update()
				b.receive <- true
			case <-b.send:
			}
		}
	}()

	return b.receive, b.send, nil
}
