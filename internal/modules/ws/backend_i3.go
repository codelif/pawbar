package ws

import (
	"sort"
	"sync"

	"github.com/codelif/pawbar/internal/services/i3"
	"github.com/codelif/pawbar/internal/utils"
)

type i3Backend struct {
	svc *i3.Service
	ev  chan interface{}
	ws  map[int]*Workspace
	mu  sync.RWMutex
	sig chan struct{}
}

func newI3Backend(s *i3.Service) backend {
	b := &i3Backend{
		svc: s,
		ev:  make(chan interface{}),
		ws:  make(map[int]*Workspace),
		sig: make(chan struct{}, 1),
	}

	b.refreshWorkspaceCache()

	b.svc.RegisterChannel("workspaces", b.ev)

	go b.loop()
	return b
}

func (b *i3Backend) loop() {
	for e := range b.ev {
		if evt, ok := e.(i3.I3Event); ok {
			utils.Logger.Println("DEBUG: ws: i3: Event type:", evt)
			b.refreshWorkspaceCache()
			b.signal()
		} else {
			utils.Logger.Println("DEBUG: ws: i3: Unknown event type", e)
		}

	}
}

func (b *i3Backend) refreshWorkspaceCache() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ws = make(map[int]*Workspace)

	workspaces := i3.GetWorkspaces()
	active := i3.GetActiveWorkspace()

	for _, w := range workspaces {
		b.ws[w.Id] = &Workspace{
			ID:     w.Id,
			Name:   w.Name,
			Active: w.Id == active.Id,
			Urgent: w.Urgent,
		}
	}
}

func (b *i3Backend) signal() {
	select {
	case b.sig <- struct{}{}:
	default:
	}
}

func (b *i3Backend) List() []Workspace {
	b.mu.RLock()
	defer b.mu.RUnlock()

	ws := make([]Workspace, 0, len(b.ws))
	for _, v := range b.ws {
		ws = append(ws, Workspace{v.ID, v.Name, v.Active, v.Urgent, v.Special})
	}
	sort.Slice(ws, func(a, b int) bool { return ws[a].ID < ws[b].ID })
	return ws
}
func (b *i3Backend) Events() <-chan struct{} { return b.sig }
func (b *i3Backend) Goto(name string)        { i3.GoToWorkspace(name) }
