package pulse

import (
	"fmt"
	"time"

	"github.com/nekorg/pawbar/internal/services"
	"github.com/codelif/pulseaudio"
)

func Register() (*PulseService, bool) {
	s, ok := services.Ensure("pulse", func() services.Service { return &PulseService{} }).(*PulseService)
	return s, ok
}

func GetService() (*PulseService, bool) {
	if s, ok := services.ServiceRegistry["pulse"].(*PulseService); ok {
		return s, true
	}
	return nil, false
}

type PulseService struct {
	running   bool
	exit      chan bool
	listeners []chan<- SinkEvent
	client    *pulseaudio.Client
}

type SinkEvent struct {
	Sink   string
	Volume float64
	Muted  bool
}

func (p *PulseService) Name() string { return "pulse" }

func (p *PulseService) IssueListener() <-chan SinkEvent {
	l := make(chan SinkEvent, 10)
	p.listeners = append(p.listeners, l)

	return l
}

func (p *PulseService) Start() error {
	if p.running {
		return nil
	}

	client, err := pulseaudio.NewClient("")
	if err != nil {
		return err
	}
	p.client = client

	events, err := client.Events()
	if err != nil {
		return err
	}

	p.exit = make(chan bool)
	p.running = true

	go func() {
		for p.running {
			select {
			case e := <-events:
				if e.Op == pulseaudio.EvChange && (e.Facility == pulseaudio.EvSink || e.Facility == pulseaudio.EvSource) {
					sink, err := p.GetDefaultSinkInfo()
					if err != nil {
						continue
					}
					for _, ch := range p.listeners {
						ch <- sink
					}
				}
			case <-p.exit:
				p.running = false
			}
		}
	}()

	return nil
}

func (p *PulseService) GetDefaultSink() (pulseaudio.Sink, error) {
	if !p.running {
		return pulseaudio.Sink{}, fmt.Errorf("pulse service not running")
	}
	serverInfo, err := p.client.ServerInfo()
	if err != nil {
		return pulseaudio.Sink{}, err
	}

	sinks, err := p.client.Sinks()
	if err != nil {
		return pulseaudio.Sink{}, err
	}

	for _, sink := range sinks {
		if sink.Name != serverInfo.DefaultSink {
			continue
		}

		return sink, nil
	}

	return pulseaudio.Sink{}, fmt.Errorf("default sink '%s' not found", serverInfo.DefaultSink)
}

func (p *PulseService) Stop() error {
	if !p.running {
		return nil
	}

	select {
	case <-time.After(2 * time.Second):
		return fmt.Errorf("could not stop")
	case p.exit <- true:
		p.running = false
	}

	return nil
}

func (p *PulseService) GetDefaultSinkInfo() (SinkEvent, error) {
	if !p.running {
		return SinkEvent{}, fmt.Errorf("pulse service not running")
	}

	sink, err := p.GetDefaultSink()
	if err != nil {
		return SinkEvent{}, err
	}

	return SinkEvent{
		Sink:   sink.Name,
		Volume: float64(float32(sink.Cvolume[0])/0xffff) * 100,
		Muted:  sink.Muted,
	}, nil
}

func (p *PulseService) SetSinkVolume(sink string, volume float64) error {
	if !p.running {
		return fmt.Errorf("pulse service not running")
	}
	return p.SetSinkVolume(sink, volume)
}

func (p *PulseService) SetSinkMute(sink string, mute bool) error {
	if !p.running {
		return fmt.Errorf("pulse service not running")
	}
	return p.SetSinkMute(sink, mute)
}
