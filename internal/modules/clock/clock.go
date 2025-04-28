package clock

import (
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/utils"
)

func New() modules.Module {
	return &ClockModule{}
}

type ClockModule struct {
	receive chan bool
	send    chan modules.Event
	format  int
}

func (c *ClockModule) Dependencies() []string {
	return nil
}

func (c *ClockModule) Update(format int) (timeFormat string) {
	if format == 2 {
		time2 := time.Now().Format("Mon 15:04")
		return time2
	}
	time1 := time.Now().Format("2006-01-02 15:04:05")
	return time1
}

func (c *ClockModule) Run() (<-chan bool, chan<- modules.Event, error) {
	c.receive = make(chan bool)
	c.send = make(chan modules.Event)
	c.format = 1

	go func() {
		t := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-t.C:
				c.receive <- true
			case e := <-c.send:
				switch ev := e.VaxisEvent.(type) {
				case vaxis.Mouse:
					utils.Logger.Printf("clock: Got mouse event: %d, %d, Mod: %d, Button: %d\n", ev.Col, ev.Row, ev.Modifiers, ev.EventType)
					// if c.format == 1 {
					// 	c.format = 2
					// } else {
					// 	c.format = 1
					// }
					// c.receive <- true
				}

			}
		}
	}()

	return c.receive, c.send, nil
}

func (c *ClockModule) Render() []modules.EventCell {
	rstring := c.Update(c.format)
	r := make([]modules.EventCell, len(rstring))
	for i, ch := range rstring {
		r[i] = modules.EventCell{
			C: vaxis.Cell{
				Character: vaxis.Character{
					Grapheme: string(ch),
					Width:    1,
				},
				Style: vaxis.Style{},
			},
			Metadata: "",
			Mod:      c,
		}
	}
	return r
}

func (c *ClockModule) Channels() (<-chan bool, chan<- modules.Event) {
	return c.receive, c.send
}

func (c *ClockModule) Name() string {
	return "clock"
}
