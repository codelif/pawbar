package pulse

import (
	"fmt"
	"time"

	"github.com/codelif/pawbar/internal/services"
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

	err := init_pulse()
	if err != nil {
		return err
	}

	events, err := monitor()
	if err != nil {
		return err
	}

	p.exit = make(chan bool)
	go func() {
		for p.running {
			select {
			case e := <-events:
				for _, ch := range p.listeners {
					ch <- e
				}
			case <-p.exit:
				p.running = false
				break
			}
		}
	}()

	return nil
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

func (p *PulseService) GetSinkInfo(sink string) (SinkInfo, error) {
	if !p.running {
		return SinkInfo{}, fmt.Errorf("pulse service not running")
	}

	return getSinkInfo(sink)
}

func (p *PulseService) GetDefaultSink() (string, error) {
	if !p.running {
		return "", fmt.Errorf("pulse service not running")
	}

	return getDefaultSink()
}
