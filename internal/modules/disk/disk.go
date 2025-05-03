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

func (d *DiskModule) Dependencies() []string {
	return nil
}

func (d *DiskModule) Run() (<-chan bool, chan<- modules.Event, error) {
	d.receive = make(chan bool)
	d.send = make(chan modules.Event)

	go func() {
		t := time.NewTicker(7 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				d.receive <- true
			case e := <-d.send:
				switch ev := e.VaxisEvent.(type) {
				case vaxis.Mouse:
					d.handleMouseEvent(ev)
				}
			}
		}
	}()
	return d.receive, d.send, nil
}

func (d *DiskModule) handleMouseEvent(ev vaxis.Mouse) {
	if ev.EventType != vaxis.EventPress {
		return
	}

	switch ev.Button {
	case vaxis.MouseLeftButton:
		d.format.toggleUnit()
		d.receive <- true
	case vaxis.MouseRightButton:
		d.format.toggleFree()
		d.receive <- true
	case vaxis.MouseMiddleButton:
		d.format = UsedPercent
		d.receive <- true
	}

}

func (d *DiskModule) formatString(du *disk.UsageStat) string {
	if du==nil{
		return ""
	}

	switch d.format {
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

func (d *DiskModule) Render() []modules.EventCell {

	du, err := disk.Usage("/")
	if err != nil {
		return nil
	}

	usage:=int(du.UsedPercent)
	s:=vaxis.Style{}
	if usage>95 {
		s.Foreground= modules.URGENT
	}else if usage>90{
		s.Foreground= modules.WARNING
	}

	icon := ''
	rch := vaxis.Characters(fmt.Sprintf("%c%s", icon, d.formatString(du)))
	r := make([]modules.EventCell, len(rch))

	for i, ch := range rch {
		r[i] = modules.EventCell{C: vaxis.Cell{Character: ch, Style: s}, Mod: d, MouseShape: vaxis.MouseShapeClickable}
	}
	return r
}

func (d *DiskModule) Channels() (<-chan bool, chan<- modules.Event) {
	return d.receive, d.send
}

func (d *DiskModule) Name() string {
	return "disk"
}
