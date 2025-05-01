package cpu

import (
	"fmt"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/shirou/gopsutil/v3/cpu"
)

func New() modules.Module {
	return &CpuModule{}
}

type CpuModule struct {
	receive chan bool
	send    chan modules.Event
}

func (c *CpuModule) Dependencies() []string {
	return nil
}

func (c *CpuModule) Run() (<-chan bool, chan<- modules.Event, error) {
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

func (c *CpuModule) Render() []modules.EventCell {
	percent, err := cpu.Percent(0, false)
	if err != nil {
		return nil
	}

	icon := ''
	rch := vaxis.Characters(fmt.Sprintf("%c %d%%", icon, int(percent[0])))
	r := make([]modules.EventCell, len(rch))

	for i, ch := range rch {
		r[i] = modules.EventCell{C: vaxis.Cell{Character: ch}, Mod: c}
	}
	return r
}

func (c *CpuModule) Channels() (<-chan bool, chan<- modules.Event) {
	return c.receive, c.send
}

func (c *CpuModule) Name() string {
	return "cpu"
}
