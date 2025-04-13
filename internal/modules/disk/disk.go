package disk

import (
	"fmt"
	"time"

	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/shirou/gopsutil/v3/disk"
)

type DiskModule struct {
	receive     chan bool
	send        chan modules.Event
	format      int
	alterformat int
	lastPref    int
	visible     string
}

func New() modules.Module {
	d := &DiskModule{
		format:      1,
		alterformat: 3,
		lastPref:    1,
	}
	d.Update(1)
	return d
}

func (d *DiskModule) Dependencies() []string {
	return nil
}

func (d *DiskModule) formatSpecifier(value int) {
	d.Update(value)
	d.lastPref = value
}

func (d *DiskModule) Update(format int) {
	du, err := disk.Usage("/")
	if err != nil {
		d.visible = ""
		return
	}
	if format == 1 {
		d.visible = fmt.Sprintf(" %d%%", int(du.UsedPercent))
	} else if format == 2 {
		d.visible = fmt.Sprintf(" %.2fGB", float64(du.Used)/1073741824.0)
	} else if format == 3 {
		d.visible = fmt.Sprintf(" %d%%", 100-int(du.UsedPercent))
	} else if format == 4 {
		d.visible = fmt.Sprintf(" %.2fGB", float64(du.Free)/1073741824.0)
	}
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
				switch ev := e.TcellEvent.(type) {
				case *tcell.EventMouse:
					if ev.Buttons() != 0 {
						x, y := ev.Position()
						btn := ev.Buttons()
						utils.Logger.Printf("Disk: Got click event: %d, %d, Mod: %d, Button: %d\n", x, y, ev.Modifiers(), btn)
						if btn == 1 {
							if d.format == 1 {
								d.format = 2
								d.formatSpecifier(2)
							} else {
								d.format = 1
								d.formatSpecifier(1)
							}
							d.receive <- true
						} else if btn == 2 {
							if d.alterformat == 3 {
								d.alterformat = 4
								d.formatSpecifier(4)
							} else {
								d.alterformat = 3
								d.formatSpecifier(3)
							}
							d.receive <- true
						}
					}
				}
			}
		}
	}()
	return d.receive, d.send, nil
}

func (d *DiskModule) Render() []modules.EventCell {
	d.Update(d.lastPref)
	icon := ''
	r := make([]modules.EventCell, len(d.visible)+1)
	r[0] = modules.EventCell{C: icon, Style: modules.DEFAULT, Metadata: "", Mod: d}
	for i, ch := range d.visible {
		r[i+1] = modules.EventCell{C: ch, Style: modules.DEFAULT, Metadata: "", Mod: d}
	}
	return r
}

func (d *DiskModule) Channels() (<-chan bool, chan<- modules.Event) {
	return d.receive, d.send
}

func (d *DiskModule) Name() string {
	return "disk"
}
