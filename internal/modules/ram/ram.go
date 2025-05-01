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

func (r *RamModule) Dependencies() []string {
	return nil
}

func (r *RamModule) Run() (<-chan bool, chan<- modules.Event, error) {
	r.receive = make(chan bool)
	r.send = make(chan modules.Event)
	r.format = 1
	go func() {
		t := time.NewTicker(3 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				r.receive <- true

			case e := <-r.send:
				switch ev := e.VaxisEvent.(type) {
				case vaxis.Mouse:
					r.handleMouseEvent(ev)
				}
			}
		}

	}()

	return r.receive, r.send, nil
}

// this is a blocking function, only use it in event loop
func (r *RamModule) handleMouseEvent(ev vaxis.Mouse) {
	if ev.EventType == vaxis.EventPress {
		switch ev.Button {
		case vaxis.MouseLeftButton:
			r.format.toggle()
			r.receive <- true
		}
	}
}

func (r *RamModule) formatString() string {
	v, err := mem.VirtualMemory()
	if err != nil {
		return ""
	}

	switch r.format {
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

func (r *RamModule) Render() []modules.EventCell {
	icon := '󰆌'

	rch := vaxis.Characters(fmt.Sprintf("%c%s", icon, r.formatString()))
	r_ := make([]modules.EventCell, len(rch))

	for i, ch := range rch {
		r_[i] = modules.EventCell{C: vaxis.Cell{Character: ch}, Mod: r, MouseShape: vaxis.MouseShapeClickable}
	}

	return r_
}

func (r *RamModule) Channels() (<-chan bool, chan<- modules.Event) {
	return r.receive, r.send
}

func (r *RamModule) Name() string {
	return "ram"
}
