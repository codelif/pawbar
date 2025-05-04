package hyprws

import (
	"errors"

	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/services/pulse"
	"gopkg.in/yaml.v3"
)

func init() {
	config.Register("volume", func(n *yaml.Node) (modules.Module, error) {
		return &VolumeModule{}, nil
	})
}

func New() modules.Module {
	return &VolumeModule{}
}

type VolumeModule struct {
	receive chan bool
	send    chan modules.Event
}

func (mod *VolumeModule) Name() string {
	return "hyprws"
}

func (mod *VolumeModule) Dependencies() []string {
	return []string{"hypr"}
}

func (mod *VolumeModule) Run() (<-chan bool, chan<- modules.Event, error) {
	service, ok := pulse.Register()
	if !ok {
		return nil, nil, errors.New("pulse service not available")
	}

	mod.receive = make(chan bool)
	mod.send = make(chan modules.Event)
	pulse_ch := service.IssueListener()

	go func() {
		for {
			select {
			case <-mod.send:
			case c := <-pulse_ch:
			}
		}
	}()

	return mod.receive, mod.send, nil
}

func (mod *VolumeModule) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *VolumeModule) Render() []modules.EventCell {
	return []modules.EventCell{}
}
