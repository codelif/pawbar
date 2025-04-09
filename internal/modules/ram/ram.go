package ram

import{
	"time"
	"fmt"
	
	"github.com/codelif/pawbar/internal/modules"
  "github.com/shirou/gopsutil/v3/mem"

}

func New() modules.Module {
	return &RAM_Module{}
}

type RAM_Module struct {
	receive chan bool
	send    chan modules.Event
}

func (r *RAM_Module) Dependencies() []string {
	return nil
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
				}
			case  <-r.send:
			}
		}
	}()

	return r.receive, r.send, nil
}

func (r *RAM_Module) Render() []modules.EventCell {
	v, err := mem.VirtualMemory()
	if err !=nil {
		return ""
	}
	icon := ''
	rstring := fmt.Sprintf(" %d%%", v.UsedPercent)
	r := make([]modules.EventCell, len(rstring)+1)
	i := 0
	r[i] = modules.EventCell{C: icon, Style: modules.DEFAULT, Metadata: "", Mod: r}
	i++
	for _, ch := range rstring {
		r[i] = modules.EventCell{C: ch, Style: modules.DEFAULT, Metadata: "", Mod: r}
		i++
	}
	return r
}


func (r *RAM_Module) Channels() (<-chan bool, chan<- modules.Event) {
	return r.receive, r.send
}

func (r *RAM_Module) Name() string {
	return "ram"
}
