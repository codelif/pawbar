package disk

import(
	"time"
	"fmt"
	"os"
	
	"github.com/codelif/pawbar/internal/modules"
  "github.com/shirou/gopsutil/v3/disk"
	"github.com/codelif/pawbar/internal/utils"
	"github.com/gdamore/tcell/v2"

)

func New() modules.Module {
	return &DiskModule{}
}

type DiskModule struct {
	receive chan bool
	send    chan modules.Event
	format  int `default : "1" `
}

func (d *DiskModule) Dependencies() []string {
	return nil
}

func (d *DiskModule) Update(format int) (value string ){
	home := os.Getenv("HOME")
	disk_, err := disk.Usage(home)
	if err !=nil {
		return ""
	}
	if format==2{
		value2:= float64(disk_.Used)/1073741824.00
		rstring := fmt.Sprintf(" %.2fGB", value2)
		return rstring
	}
	value1:= int(disk_.UsedPercent)
	rstring := fmt.Sprintf(" %d%%", value1)
	return rstring
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
		
			case e:=  <-d.send:
			switch ev := e.TcellEvent.(type) {
			case *tcell.EventMouse:
				if ev.Buttons() != 0 {
					x, y := ev.Position()
					utils.Logger.Printf("Clock: Got click event: %d, %d, Mod: %d, Button: %d\n", x, y, ev.Modifiers(), ev.Buttons())
					if d.format==1 {
						d.format=2
					}else{
						d.format=1
					}
				}
			}
			}
			}
		
	}()

	return d.receive, d.send, nil
}

func (d *DiskModule) Render() []modules.EventCell {
	icon := ''
	rstring:= d.Update(d.format)
	r_ := make([]modules.EventCell, len(rstring)+1)
	i := 0
	r_[i] = modules.EventCell{C: icon, Style: modules.DEFAULT, Metadata: "", Mod: d}
	i++
	for _, ch := range rstring {
		r_[i] = modules.EventCell{C: ch, Style: modules.DEFAULT, Metadata: "", Mod: d}
		i++
	}
	return r_
}


func (d *DiskModule) Channels() (<-chan bool, chan<- modules.Event) {
	return d.receive, d.send
}

func (d *DiskModule) Name() string {
	return "disk"
}
