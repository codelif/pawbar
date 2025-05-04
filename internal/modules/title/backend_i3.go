package title

import (
	"github.com/codelif/pawbar/internal/services/i3"
	"github.com/codelif/pawbar/internal/utils"
)

type i3Backend struct {
	svc      *i3.Service
	ev       chan interface{}
	ev2      chan interface{}
	instance string
	title    string
	sig      chan struct{}
}

func newI3Backend(s *i3.Service) backend {
	b := &i3Backend{
		svc: s,
		ev:  make(chan interface{}),
		ev2: make(chan interface{}),
		sig: make(chan struct{}, 2),
	}

	b.instance, b.title = i3.GetTitleClass()

	b.svc.RegisterChannel("activeWindow", b.ev)
	b.svc.RegisterChannel("workspaces", b.ev2)

	go b.loop()
	return b
}

func (b *i3Backend) loop() {
	for {
		select {
		case e := <-b.ev:
			if _, ok := e.(i3.I3WEvent); ok {
				b.instance, b.title = i3.GetTitleClass()
				b.signal()
			} else {
        utils.Logger.Println("DEBUG: ws: i3: Unknown event on window event channel:", e)
			}
    case e:= <-b.ev2:
			if _, ok := e.(i3.I3Event); ok {
				b.instance, b.title = i3.GetTitleClass()
				b.signal()
			} else {
        utils.Logger.Println("DEBUG: ws: i3: Unknown event type on workspace event channel:", e)
			}
		}
	}
}

func (b *i3Backend) signal() {
	select {
	case b.sig <- struct{}{}:
	default:
	}
}

func (b *i3Backend) Window() Window {
	return Window{Title: b.title, Class: b.instance}
}
func (b *i3Backend) Events() <-chan struct{} { return b.sig }
