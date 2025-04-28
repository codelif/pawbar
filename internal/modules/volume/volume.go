package hyprws

import (
	"errors"

	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/services/pulse"
)

func New() modules.Module {
	return &VolumeModule{}
}

type VolumeModule struct {
	receive chan bool
	send    chan modules.Event
}

func (vol *VolumeModule) Name() string {
	return "hyprws"
}

func (vol *VolumeModule) Dependencies() []string {
	return []string{"hypr"}
}

func (vol *VolumeModule) Run() (<-chan bool, chan<- modules.Event, error) {
	service, ok := pulse.GetService()
	if !ok {
		return nil, nil, errors.New("pulse service not available")
	}

	vol.receive = make(chan bool)
	vol.send = make(chan modules.Event)
	pulse_ch := service.IssueListener()

	go func() {
		for {
			select {
			case <-vol.send:
			case c := <-pulse_ch:
			}
		}
	}()

	return vol.receive, vol.send, nil
}

func (vol *VolumeModule) Channels() (<-chan bool, chan<- modules.Event) {
	return vol.receive, vol.send
}

func (vol *VolumeModule) Render() []modules.EventCell {
	return []modules.EventCell{}
}
