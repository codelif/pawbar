package ram

import (
	"fmt"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/shirou/gopsutil/v3/mem"
)

func New() modules.Module {
	return &RamModule{}
}

type Format int

const (
	FormatPercentage Format = iota
	FormatAbsolute
)

func (f *Format) toggle() { *f ^= 1 }

type RamModule struct {
	receive chan bool
	send    chan modules.Event
	format  Format
}

func (mod *RamModule) Dependencies() []string {
	return nil
}

func (mod *RamModule) Run() (<-chan bool, chan<- modules.Event, error) {
	mod.receive = make(chan bool)
	mod.send = make(chan modules.Event)
	go func() {
		t := time.NewTicker(3 * time.Second)
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

// this is a blocking function, only use it in event loop
func (mod *RamModule) handleMouseEvent(ev vaxis.Mouse) {
	if ev.EventType == vaxis.EventPress {
		switch ev.Button {
		case vaxis.MouseLeftButton:
			mod.format.toggle()
			mod.receive <- true
		}
	}
}

func (mod *RamModule) formatString(v *mem.VirtualMemoryStat) string {
	if v == nil {
		return ""
	}
	switch mod.format {
	case FormatPercentage:
		value := int(v.UsedPercent)
		rstring := fmt.Sprintf(" %d%%", value)
		return rstring
	case FormatAbsolute:
		value := float64(v.Used) / 1073741824.00
		rstring := fmt.Sprintf(" %.2fGB", value)
		return rstring
	}
	return ""
}

func (mod *RamModule) Render() []modules.EventCell {
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil
	}

	thresValue := int(v.UsedPercent)
	s := vaxis.Style{}
	if thresValue > 90 {
		s.Foreground = modules.URGENT
	} else if thresValue > 80 {
		s.Foreground = modules.WARNING
	}

	icon := '󰆌'

	rch := vaxis.Characters(fmt.Sprintf("%c%s", icon, mod.formatString(v)))
	r_ := make([]modules.EventCell, len(rch))

	for i, ch := range rch {
		r_[i] = modules.EventCell{C: vaxis.Cell{Character: ch, Style: s}, Mod: mod, MouseShape: vaxis.MouseShapeClickable}
	}

	return r_
}

func (mod *RamModule) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *RamModule) Name() string {
	return "ram"
}
