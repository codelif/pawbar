package disk

import (
	"fmt"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/shirou/gopsutil/v3/disk"
)

type DiskModule struct {
	receive chan bool
	send    chan modules.Event
	format  Format
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

func New() modules.Module {
	return &DiskModule{}
}

func (mod *DiskModule) Dependencies() []string {
	return nil
}

func (mod *DiskModule) Run() (<-chan bool, chan<- modules.Event, error) {
	mod.receive = make(chan bool)
	mod.send = make(chan modules.Event)

	go func() {
		t := time.NewTicker(7 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				mod.receive <- true
			case e := <-mod.send:
				switch ev := e.VaxisEvent.(type) {
				case vaxis.Mouse:
					mod.handleMouseEvent(ev)
				}
			}
		}
	}()
	return mod.receive, mod.send, nil
}

func (mod *DiskModule) handleMouseEvent(ev vaxis.Mouse) {
	if ev.EventType != vaxis.EventPress {
		return
	}

	switch ev.Button {
	case vaxis.MouseLeftButton:
		mod.format.toggleUnit()
		mod.receive <- true
	case vaxis.MouseRightButton:
		mod.format.toggleFree()
		mod.receive <- true
	case vaxis.MouseMiddleButton:
		mod.format = UsedPercent
		mod.receive <- true
	}
}

func (mod *DiskModule) formatString(du *disk.UsageStat) string {
	if du == nil {
		return ""
	}

	switch mod.format {
	case UsedPercent:
		return fmt.Sprintf(" %d%%", int(du.UsedPercent))
	case UsedAbsolute:
		return fmt.Sprintf(" %.2fGB", float64(du.Used)/GiB)
	case FreePercent:
		return fmt.Sprintf(" %d%%", 100-int(du.UsedPercent))
	case FreeAbsolute:
		return fmt.Sprintf(" %.2fGB", float64(du.Free)/GiB)
	}

	return ""
}

func (mod *DiskModule) Render() []modules.EventCell {
	du, err := disk.Usage("/")
	if err != nil {
		return nil
	}

	usage := int(du.UsedPercent)
	s := vaxis.Style{}
	if usage > 95 {
		s.Foreground = modules.URGENT
	} else if usage > 90 {
		s.Foreground = modules.WARNING
	}

	icon := ''
	rch := vaxis.Characters(fmt.Sprintf("%c%s", icon, mod.formatString(du)))
	r := make([]modules.EventCell, len(rch))

	for i, ch := range rch {
		r[i] = modules.EventCell{C: vaxis.Cell{Character: ch, Style: s}, Mod: mod, MouseShape: vaxis.MouseShapeClickable}
	}
	return r
}

func (mod *DiskModule) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *DiskModule) Name() string {
	return "disk"
}
