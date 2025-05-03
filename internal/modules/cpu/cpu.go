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
	receive       chan bool
	send          chan modules.Event
	highStart     time.Time
	highTriggered bool
	required      time.Duration
}

func (mod *CpuModule) Dependencies() []string {
	return nil
}

func (mod *CpuModule) Run() (<-chan bool, chan<- modules.Event, error) {
	mod.receive = make(chan bool)
	mod.send = make(chan modules.Event)

	go func() {
		t := time.NewTicker(3 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				mod.receive <- true
			case <-mod.send:
			}
		}
	}()

	return mod.receive, mod.send, nil
}

func (mod *CpuModule) Render() []modules.EventCell {
	percent, err := cpu.Percent(0, false)
	if err != nil {
		return nil
	}
	usage := int(percent[0])
	const threshold = 90

	if usage > threshold {
		if mod.highStart.IsZero() {
			mod.highStart = time.Now()
		} else if !mod.highTriggered && time.Since(mod.highStart) >= mod.required {
			mod.highTriggered = true
		}
	} else {
		mod.highStart = time.Time{}
		mod.highTriggered = false
	}

	s := vaxis.Style{}
	if mod.highTriggered {
		s.Foreground = modules.URGENT
	}

	icon := ''
	rch := vaxis.Characters(fmt.Sprintf("%c %d%%", icon, usage))
	r := make([]modules.EventCell, len(rch))

	for i, ch := range rch {
		r[i] = modules.EventCell{C: vaxis.Cell{Character: ch, Style: s}, Mod: mod}
	}
	return r
}

func (mod *CpuModule) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *CpuModule) Name() string {
	return "cpu"
}
