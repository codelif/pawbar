package cpu

import(
	"time"
	"fmt"
	
	"github.com/codelif/pawbar/internal/modules"
	"github.com/shirou/gopsutil/v3/cpu"
)

func New() modules.Module {
	return &CPU_Module{}
}

type CPU_Module struct {
	receive chan bool
	send    chan modules.Event
}

func (c *CPU_Module) Dependencies() []string {
	return nil
}

func (c *CPU_Module) Run() (<-chan bool, chan<- modules.Event, error) {
	c.receive = make(chan bool)
	c.send = make(chan modules.Event)

	go func() {
		t := time.NewTicker(3 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
					c.receive <- true
			case <-c.send:
			}
			}
	}()

	return c.receive, c.send, nil
}

func (c *CPU_Module) Render() []modules.EventCell {
	percent, err := cpu.Percent(0, false)
	if err !=nil {
		return nil
	}
	icon := ''
	rstring := fmt.Sprintf(" %d%%", int(percent[0]))
	r := make([]modules.EventCell, len(rstring)+1)
	i := 0
	r[i] = modules.EventCell{C: icon, Style: modules.DEFAULT, Metadata: "", Mod: c}
	i++
	for _, ch := range rstring {
		r[i] = modules.EventCell{C: ch, Style: modules.DEFAULT, Metadata: "", Mod: c}
		i++
	}
	return r
}


func (c *CPU_Module) Channels() (<-chan bool, chan<- modules.Event) {
	return c.receive, c.send
}

func (c *CPU_Module)  Name() string {
	return "cpu"
}
