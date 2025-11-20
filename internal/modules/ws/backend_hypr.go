package ws

import (
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/nekorg/pawbar/internal/services/hypr"
)

type hyprBackend struct {
	svc *hypr.Service
	ev  chan hypr.HyprEvent
	ws  map[int]*Workspace
	mu  sync.RWMutex
	sig chan struct{}
}

func newHyprBackend(s *hypr.Service) backend {
	b := &hyprBackend{
		svc: s,
		ev:  make(chan hypr.HyprEvent),
		ws:  make(map[int]*Workspace),
		sig: make(chan struct{}, 1),
	}

	b.refreshWorkspaceCache()

	for _, e := range []string{"workspacev2", "focusedmonv2", "createworkspacev2", "destroyworkspacev2", "activespecial", "renameworkspace", "urgent"} {
		b.svc.RegisterChannel(e, b.ev)
	}

	go b.loop()
	return b
}

func (b *hyprBackend) loop() {
	for e := range b.ev {
		if !b.validate(e) {
			b.refreshWorkspaceCache()
		}

		if b.handleEvent(e) {
			b.signal()
		}
	}
}

func (b *hyprBackend) refreshWorkspaceCache() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ws = make(map[int]*Workspace)

	workspaces := hypr.GetWorkspaces()

	for _, w := range workspaces {
		b.ws[w.Id] = &Workspace{
			ID:      w.Id,
			Name:    w.Name,
			Active:  w.Id == hypr.GetActiveWorkspace().Id,
			Special: strings.HasPrefix(w.Name, "special:"),
		}
	}
}

func (b *hyprBackend) signal() {
	select {
	case b.sig <- struct{}{}:
	default:
	}
}

func (b *hyprBackend) validate(e hypr.HyprEvent) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch e.Event {
	case "workspacev2":
		id, _ := strconv.Atoi(e.Data[:strings.IndexRune(e.Data, ',')])
		_, ok := b.ws[id]
		return ok

	case "focusedmonv2":
		id, _ := strconv.Atoi(e.Data[strings.LastIndex(e.Data, ",")+1:])
		_, ok := b.ws[id]
		return ok

	case "createworkspacev2":
		id, _ := strconv.Atoi(e.Data[:strings.IndexRune(e.Data, ',')])
		_, ok := b.ws[id]
		return !ok

	case "destroyworkspacev2":
		id, _ := strconv.Atoi(e.Data[:strings.IndexRune(e.Data, ',')])
		_, ok := b.ws[id]
		return ok

	case "renameworkspace":
		id, _ := strconv.Atoi(e.Data[:strings.IndexRune(e.Data, ',')])
		_, ok := b.ws[id]
		return ok
	}

	return true
}

func (b *hyprBackend) handleEvent(e hypr.HyprEvent) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch e.Event {
	case "workspacev2":
		id, _ := strconv.Atoi(e.Data[:strings.IndexRune(e.Data, ',')])
		b.setActiveWorkspace(id)
	case "createworkspacev2":
		id_str, name, _ := strings.Cut(e.Data, ",")
		id, _ := strconv.Atoi(id_str)
		b.createWorkspace(id, name)
	case "destroyworkspacev2":
		id, _ := strconv.Atoi(e.Data[:strings.IndexRune(e.Data, ',')])
		b.destroyWorkspace(id)
	case "activespecial":
		b.activateSpecialWorkspace(e.Data[:strings.IndexRune(e.Data, ',')])
	case "urgent":
		b.setWorkspaceUrgent(e.Data)
	case "renameworkspace":
		idr, name, _ := strings.Cut(e.Data, ",")
		id, _ := strconv.Atoi(idr)
		b.renameWorkspace(id, name)
	default:
		return false
	}
	return true
}

func (b *hyprBackend) renameWorkspace(id int, name string) {
	for _, w := range b.ws {
		if w.ID == id {
			w.Name = name
		}
	}
}

func (b *hyprBackend) setActiveWorkspace(id int) {
	for _, w := range b.ws {
		if !w.Special {
			w.Active = false
		}
	}

	b.ws[id].Active = true
	b.ws[id].Urgent = false
}

func (b *hyprBackend) createWorkspace(id int, name string) {
	b.ws[id] = &Workspace{
		ID:      id,
		Name:    name,
		Active:  false,
		Special: strings.HasPrefix(name, "special:"),
	}
}

func (b *hyprBackend) destroyWorkspace(id int) {
	delete(b.ws, id)
}

func (b *hyprBackend) activateSpecialWorkspace(name string) {
	active := name != ""

	for _, w := range b.ws {
		if w.Special {
			w.Active = active
		}
	}
}

func (b *hyprBackend) setWorkspaceUrgent(address string) {
	clients := hypr.GetClients()

	activeId := 0
	for _, w := range b.ws {
		if w.Active && !w.Special {
			activeId = w.ID
		}
	}

	for _, client := range clients {
		client_address, _ := strings.CutPrefix(client.Address, "0x")
		if client_address == address && client.Workspace.Id != activeId {
			b.ws[client.Workspace.Id].Urgent = true
		}
	}
}

func (b *hyprBackend) List() []Workspace {
	b.mu.RLock()
	defer b.mu.RUnlock()
	ws := make([]Workspace, 0, len(b.ws))
	for _, v := range b.ws {
		ws = append(ws, Workspace{v.ID, v.Name, v.Active, v.Urgent, v.Special})
	}
	sort.Slice(ws, func(a, b int) bool { return ws[a].ID < ws[b].ID })
	return ws
}
func (b *hyprBackend) Events() <-chan struct{} { return b.sig }
func (b *hyprBackend) Goto(name string)        { hypr.GoToWorkspace(name) }
