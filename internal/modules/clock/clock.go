package clock

import (
	"time"

	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/utils"
	"github.com/gdamore/tcell/v2"
)

func New() modules.Module {
	return &ClockModule{}
}

type ClockModule struct {
	receive chan bool
	send    chan modules.Event
}

func (c *ClockModule) Dependencies() []string {
	return nil
}

func (c *ClockModule) Run() (<-chan bool, chan<- modules.Event, error) {
	c.receive = make(chan bool)
	c.send = make(chan modules.Event)

	go func() {
		t := time.NewTicker(time.Second)
		for {
			select {
			case <-t.C:
				c.receive <- true
			case e := <-c.send:
				switch ev := e.TcellEvent.(type) {
				case *tcell.EventMouse:
					if ev.Buttons() != 0 {
						x, y := ev.Position()
						utils.Logger.Printf("Clock: Got click event: %d, %d, Mod: %d, Button: %d\n", x, y, ev.Modifiers(), ev.Buttons())
					}
				}
			}
		}
	}()

	return c.receive, c.send, nil
}

func (c *ClockModule) Render() []modules.EventCell {
	rstring := time.Now().Format("2006-01-02 15:04:05")
	r := make([]modules.EventCell, len(rstring))
	for i, ch := range rstring {
		r[i] = modules.EventCell{C: ch, Style: modules.DEFAULT, Metadata: "", Mod: c}
	}

	return r
}

func (c *ClockModule) Channels() (<-chan bool, chan<- modules.Event) {
	return c.receive, c.send
}

func (c *ClockModule) Name() string {
	return "clock"
}
