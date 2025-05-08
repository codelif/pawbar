package volume

import (
	"errors"
	"fmt"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/services/pulse"
	"gopkg.in/yaml.v3"
)

type VolumeModule struct {
	receive chan bool
	send    chan modules.Event
	svc     *pulse.PulseService
	sink    string
	Volume  float64
	Muted   bool
	events  <-chan pulse.SinkEvent
}

var (
	ICONS_VOLUME      = []rune{'󰕿', '󰖀', '󰕾'}
	MUTE         rune = '󰖁'
)

func init() {
	config.Register("volume", func(n *yaml.Node) (modules.Module, error) {
		return &VolumeModule{}, nil
	})
}

func New() modules.Module {
	return &VolumeModule{}
}

func (mod *VolumeModule) Name() string {
	return "volume"
}

func (mod *VolumeModule) Dependencies() []string {
	return []string{"pulse"}
}

func (mod *VolumeModule) Run() (<-chan bool, chan<- modules.Event, error) {
	svc, ok := pulse.Register()
	if !ok {
		return nil, nil, errors.New("pulse service not available")
	}
	if err := svc.Start(); err != nil {
		return nil, nil, err
	}

	sinkName, err := svc.GetDefaultSink()
	if err != nil {
		return nil, nil, err
	}
	info, err := svc.GetSinkInfo(sinkName)
	if err != nil {
		return nil, nil, err
	}

	mod.svc = svc
	mod.sink = sinkName
	mod.Volume = info.Volume
	mod.Muted = info.Muted

	mod.receive = make(chan bool)
	mod.send = make(chan modules.Event)
	mod.events = svc.IssueListener()

	go func() {
		for {
			select {
			case <-mod.send:
			case e := <-mod.events:
				if e.Sink != mod.sink {
					continue
				}
				mod.Volume = e.Volume
				mod.Muted = e.Muted
				mod.receive <- true
			}
		}
	}()

	return mod.receive, mod.send, nil
}

// in case of tui control scollup/down etc

// func (mod *VolumeModule) Change(delta float64) error {
// 	v := mod.Volume + delta
// 	if v < 0 {
// 		v = 0
// 	}
// 	if v > 1 {
// 		v = 1
// 	}
// 	if err := mod.svc.SetSinkVolume(mod.sink, v); err != nil {
// 		return err
// 	}
// 	mod.Volume = v
// 	return nil
// }
//
// func (mod *VolumeModule) ToggleMute() error {
// 	newMute := !mod.Muted
// 	if err := mod.svc.SetSinkMute(mod.sink, newMute); err != nil {
// 		return err
// 	}
// 	mod.Muted = newMute
// 	return nil
// }

func (mod *VolumeModule) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *VolumeModule) Render() []modules.EventCell {
	if mod.Muted {
		rch := vaxis.Characters(fmt.Sprintf("%c %s", MUTE, "MUTED"))
		r := make([]modules.EventCell, len(rch))
		s := vaxis.Style{}
		s.Foreground = vaxis.RGBColor(169, 169, 169)
		for i, ch := range rch {
			r[i] = modules.EventCell{C: vaxis.Cell{Character: ch, Style: s}, Mod: mod}
		}
		return r
	} else {
		vol := int(mod.Volume)
		idx := (len(ICONS_VOLUME) - 1) * vol / 100
		icon := ICONS_VOLUME[idx]
		rch := vaxis.Characters(fmt.Sprintf("%c %d%%", icon, vol))
		r := make([]modules.EventCell, len(rch))
		for i, ch := range rch {
			r[i] = modules.EventCell{C: vaxis.Cell{Character: ch}, Mod: mod}
		}
		return r

	}
}
