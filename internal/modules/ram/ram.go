package ram

import(
	"time"
	"fmt"
	
	"github.com/codelif/pawbar/internal/modules"
  "github.com/shirou/gopsutil/v3/mem"
	"github.com/codelif/pawbar/internal/utils"
	"github.com/gdamore/tcell/v2"

)

func New() modules.Module {
	return &RAM_Module{}
}

type RAM_Module struct {
	receive chan bool
	send    chan modules.Event
	format  int `default : "1" `
}

func (r *RAM_Module) Dependencies() []string {
	return nil
}

func (r* RAM_Module) Update(format int) (value string ){
	v, err := mem.VirtualMemory()
	if err !=nil {
		return ""
	}
	if format==2{
		value2:= float64(v.Used)/1073741824.00
		rstring := fmt.Sprintf(" %.2fGB", value2)
		return rstring
	}
	value1:= int(v.UsedPercent)
	rstring := fmt.Sprintf(" %d%%", value1)
	return rstring
}


func (r *RAM_Module) Run() (<-chan bool, chan<- modules.Event, error) {
	r.receive = make(chan bool)
	r.send = make(chan modules.Event)

	go func() {
		t := time.NewTicker(3 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
					r.receive <- true
		
			case e:=  <-r.send:
			switch ev := e.TcellEvent.(type) {
			case *tcell.EventMouse:
				if ev.Buttons() != 0 {
					x, y := ev.Position()
					utils.Logger.Printf("Clock: Got click event: %d, %d, Mod: %d, Button: %d\n", x, y, ev.Modifiers(), ev.Buttons())
					if r.format==1 {
						r.format=2
					}else{
						r.format=1
					}
				}
			}
			}
			}
		
	}()

	return r.receive, r.send, nil
}

func (r *RAM_Module) Render() []modules.EventCell {
	icon := '󰆌'
	rstring:= r.Update(r.format)
	r_ := make([]modules.EventCell, len(rstring)+1)
	i := 0
	r_[i] = modules.EventCell{C: icon, Style: modules.DEFAULT, Metadata: "", Mod: r}
	i++
	for _, ch := range rstring {
		r_[i] = modules.EventCell{C: ch, Style: modules.DEFAULT, Metadata: "", Mod: r}
		i++
	}
	return r_
}


func (r *RAM_Module) Channels() (<-chan bool, chan<- modules.Event) {
	return r.receive, r.send
}

func (r *RAM_Module) Name() string {
	return "ram"
}
