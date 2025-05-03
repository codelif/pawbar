package hyprws

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"sync"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/services/hypr"
)

type WorkspaceState struct {
	id     int
	name   string
	active bool
	urgent bool
}

func New() modules.Module {
	return &HyprWorkspaceModule{}
}

type HyprWorkspaceModule struct {
	receive   chan bool
	send      chan modules.Event
	hevent    chan hypr.HyprEvent
	ws        map[int]*WorkspaceState
	activeId  int
	specialId int
	special   bool
	mu        sync.Mutex
}

func (mod *HyprWorkspaceModule) Name() string {
	return "hyprws"
}

func (mod *HyprWorkspaceModule) Dependencies() []string {
	return []string{"hypr"}
}

func (mod *HyprWorkspaceModule) Run() (<-chan bool, chan<- modules.Event, error) {
	service, ok := hypr.Register()
	if !ok {
		return nil, nil, errors.New("Hypr service not available")
	}

	mod.receive = make(chan bool)
	mod.send = make(chan modules.Event)
	mod.hevent = make(chan hypr.HyprEvent)
	for _, e := range []string{"workspacev2", "focusedmonv2", "createworkspacev2", "destroyworkspacev2", "activespecial", "renameworkspace", "urgent"} {
		service.RegisterChannel(e, mod.hevent)
	}
	mod.refreshWorkspaceCache()
	go func() {
		for {
			select {
			case e := <-mod.send:
				switch ev := e.VaxisEvent.(type) {
				case vaxis.Mouse:
					mod.handleMouseEvent(e, ev)
				}
			case h := <-mod.hevent:
				if !mod.validateHyprEvent(h) {
					mod.refreshWorkspaceCache()
				}

				if mod.handleHyprEvent(h) {
					mod.receive <- true
				}
			}
		}
	}()

	return mod.receive, mod.send, nil
}

func (mod *HyprWorkspaceModule) handleMouseEvent(e modules.Event, ev vaxis.Mouse) {
	if ev.EventType != vaxis.EventPress {
		return
	}

	switch ev.Button {
	case vaxis.MouseLeftButton:
		go hypr.GoToWorkspace(e.Cell.Metadata)
	}
}

func (mod *HyprWorkspaceModule) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *HyprWorkspaceModule) refreshWorkspaceCache() {
	mod.mu.Lock()
	defer mod.mu.Unlock()
	mod.ws = make(map[int]*WorkspaceState)

	workspaces := hypr.GetWorkspaces()
	active := hypr.GetActiveWorkspace()

	for _, ws := range workspaces {
		mod.ws[ws.Id] = &WorkspaceState{ws.Id, ws.Name, ws.Id == active.Id, false}
		if ws.Name == "special:magic" {
			mod.specialId = ws.Id
			mod.special = true
		}
		if ws.Id == active.Id {
			mod.activeId = ws.Id
		}
	}
}

func (mod *HyprWorkspaceModule) validateHyprEvent(e hypr.HyprEvent) bool {
	switch e.Event {
	case "workspacev2":
		id, _ := strconv.Atoi(e.Data[:strings.IndexRune(e.Data, ',')])
		_, ok := mod.ws[id]
		return ok

	case "focusedmonv2":
		id, _ := strconv.Atoi(e.Data[strings.LastIndex(e.Data, ",")+1:])
		_, ok := mod.ws[id]
		return ok

	case "createworkspacev2":
		id, _ := strconv.Atoi(e.Data[:strings.IndexRune(e.Data, ',')])
		_, ok := mod.ws[id]
		return !ok

	case "destroyworkspacev2":
		id, _ := strconv.Atoi(e.Data[:strings.IndexRune(e.Data, ',')])
		_, ok := mod.ws[id]
		return ok

	case "renameworkspace":
		id, _ := strconv.Atoi(e.Data[:strings.IndexRune(e.Data, ',')])
		_, ok := mod.ws[id]
		return ok
	}

	return true
}

func (mod *HyprWorkspaceModule) setActiveWorkspace(id int) {
	mod.ws[mod.activeId].active = false
	mod.ws[id].active = true
	mod.ws[id].urgent = false
	mod.activeId = id
}

func (mod *HyprWorkspaceModule) createWorkspace(id int, name string) {
	mod.ws[id] = &WorkspaceState{id, name, false, false}
	if name == "special:magic" {
		mod.specialId = id
		mod.special = true
	}
}

func (mod *HyprWorkspaceModule) destroyWorkspace(id int) {
	delete(mod.ws, id)
	if id == mod.specialId {
		mod.special = false
	}
}

func (mod *HyprWorkspaceModule) isSpecialWorkspaceActive() bool {
	if mod.special {
		return mod.ws[mod.specialId].active
	}
	return false
}

func (mod *HyprWorkspaceModule) activateSpecialWorkspace(name string) {
	if name == "" {
		mod.ws[mod.specialId].active = false
	} else {
		mod.ws[mod.specialId].active = true
	}
}

func (mod *HyprWorkspaceModule) setWorkspaceUrgent(address string) {
	clients := hypr.GetClients()
	for _, client := range clients {
		client_address, _ := strings.CutPrefix(client.Address, "0x")
		if client_address == address && client.Workspace.Id != mod.activeId {
			mod.ws[client.Workspace.Id].urgent = true
		}
	}
}

func (mod *HyprWorkspaceModule) handleHyprEvent(e hypr.HyprEvent) bool {
	mod.mu.Lock()
	defer mod.mu.Unlock()
	switch e.Event {
	case "workspacev2":
		id, _ := strconv.Atoi(e.Data[:strings.IndexRune(e.Data, ',')])
		mod.setActiveWorkspace(id)
	case "createworkspacev2":
		id_str, name, _ := strings.Cut(e.Data, ",")
		id, _ := strconv.Atoi(id_str)
		mod.createWorkspace(id, name)
	case "destroyworkspacev2":
		id, _ := strconv.Atoi(e.Data[:strings.IndexRune(e.Data, ',')])
		mod.destroyWorkspace(id)
	case "activespecial":
		mod.activateSpecialWorkspace(e.Data[:strings.IndexRune(e.Data, ',')])
	case "urgent":
		mod.setWorkspaceUrgent(e.Data)
	default:
		return false
	}
	return true
}

var (
	SPECIAL = vaxis.Style{Foreground: modules.ACTIVE, Background: modules.SPECIAL}
	ACTIVE  = vaxis.Style{Foreground: modules.BLACK, Background: modules.ACTIVE}
	URGENT  = vaxis.Style{Foreground: modules.BLACK, Background: modules.URGENT}
)

func (mod *HyprWorkspaceModule) Render() []modules.EventCell {
	mod.mu.Lock()
	defer mod.mu.Unlock()
	var wss []int
	for k := range mod.ws {
		if k > 0 {
			wss = append(wss, k)
		}
	}

	slices.Sort(wss)

	var r []modules.EventCell

	if mod.isSpecialWorkspaceActive() {
		for _, ch := range vaxis.Characters(" S ") {
			r = append(r, modules.EventCell{C: vaxis.Cell{Character: ch, Style: SPECIAL}, Metadata: mod.ws[mod.specialId].name, Mod: mod})
		}
	}

	for _, id := range wss {
		wsName := mod.ws[id].name
		style := vaxis.Style{}
		mouseShape := vaxis.MouseShapeClickable

		if mod.ws[id].active {
			style = ACTIVE
			mouseShape = vaxis.MouseShapeDefault
		} else if mod.ws[id].urgent {
			style = URGENT
		}

		for _, ch := range vaxis.Characters(" " + wsName + " ") {
			r = append(r, modules.EventCell{C: vaxis.Cell{Character: ch, Style: style}, Metadata: wsName, Mod: mod, MouseShape: mouseShape})
		}
	}

	return r
}
