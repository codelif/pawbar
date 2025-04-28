package clock

import (
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/modules"
)

func New() modules.Module {
	return &ClockModule{}
}

type Format int

const (
	FormatDefault = iota
	FormatAlt1
)

type ClockModule struct {
	receive chan bool
	send    chan modules.Event
	format  Format
}

func (c *ClockModule) Dependencies() []string {
	return nil
}


func (c *ClockModule) Run() (<-chan bool, chan<- modules.Event, error) {
	c.receive = make(chan bool)
	c.send = make(chan modules.Event)

	go func() {
		t := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-t.C:
				c.receive <- true
			case e := <-c.send:
				switch ev := e.VaxisEvent.(type) {
				case vaxis.Mouse:
					c.handleMouseEvent(ev)
				}

			}
		}
	}()

	return c.receive, c.send, nil
}

func (c *ClockModule) timeFormatString() string {
	switch c.format {
	case FormatDefault:
		return time.Now().Format("2006-01-02 15:04:05")
	case FormatAlt1:
		return time.Now().Format("Mon 15:04")
	}
	return time.Now().Format("2006-01-02 15:04:05")
}


// this is a blocking function, only use it in event loop
func (c *ClockModule) handleMouseEvent(ev vaxis.Mouse) {
	if ev.EventType == vaxis.EventPress {
		switch ev.Button {
		case vaxis.MouseLeftButton:
			c.cycle()
			c.receive <- true
		}
	}
}

func (c *ClockModule) cycle() {
	switch c.format {
	case FormatDefault:
		c.format = FormatAlt1
	case FormatAlt1:
		c.format = FormatDefault
	}
}

func (c *ClockModule) Render() []modules.EventCell {
	rstring := c.timeFormatString()
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
