package hyprws

import (
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/services/hypr"
	"github.com/gdamore/tcell/v2"
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
}

func (wsMod *HyprWorkspaceModule) Name() string {
	return "hyprws"
}

func (wsMod *HyprWorkspaceModule) Dependencies() []string {
	return []string{"hypr"}
}

func (wsMod *HyprWorkspaceModule) Run() (<-chan bool, chan<- modules.Event, error) {
	service, ok := hypr.GetService()
	if !ok {
		return nil, nil, errors.New("Hypr service not available")
	}

	wsMod.receive = make(chan bool)
	wsMod.send = make(chan modules.Event)
	wsMod.hevent = make(chan hypr.HyprEvent)
	for _, e := range []string{"workspacev2", "focusedmonv2", "createworkspacev2", "destroyworkspacev2", "activespecial", "renameworkspace", "urgent"} {
		service.RegisterChannel(e, wsMod.hevent)
	}
	wsMod.refreshWorkspaceCache()
	go func() {
		for {
			select {
			case e := <-wsMod.send:
				switch ev := e.TcellEvent.(type) {
				case *tcell.EventMouse:
					btns := ev.Buttons()
					if btns == tcell.Button1 {
						go hypr.GoToWorkspace(e.Cell.Metadata)
					}
				}
			case h := <-wsMod.hevent:
				if !wsMod.validateHyprEvent(h) {
					wsMod.refreshWorkspaceCache()
				}
				if wsMod.handleHyprEvent(h) {
					wsMod.receive <- true
				}
			}
		}
	}()

	return wsMod.receive, wsMod.send, nil
}

func (wsMod *HyprWorkspaceModule) Channels() (<-chan bool, chan<- modules.Event) {
	return wsMod.receive, wsMod.send
}

func (wsMod *HyprWorkspaceModule) refreshWorkspaceCache() {
	wsMod.ws = make(map[int]*WorkspaceState)

	workspaces := hypr.GetWorkspaces()
	active := hypr.GetActiveWorkspace()

	for _, ws := range workspaces {
		wsMod.ws[ws.Id] = &WorkspaceState{ws.Id, ws.Name, ws.Id == active.Id, false}
		if ws.Name == "special:magic" {
			wsMod.specialId = ws.Id
			wsMod.special = true
		}
		if ws.Id == active.Id {
			wsMod.activeId = ws.Id
		}
	}
}

func (wsMod *HyprWorkspaceModule) validateHyprEvent(e hypr.HyprEvent) bool {
	switch e.Event {
	case "workspacev2":
		id, _ := strconv.Atoi(e.Data[:strings.IndexRune(e.Data, ',')])
		_, ok := wsMod.ws[id]
		return ok

	case "focusedmonv2":
		id, _ := strconv.Atoi(e.Data[strings.LastIndex(e.Data, ",")+1:])
		_, ok := wsMod.ws[id]
		return ok

	case "createworkspacev2":
		id, _ := strconv.Atoi(e.Data[:strings.IndexRune(e.Data, ',')])
		_, ok := wsMod.ws[id]
		return !ok

	case "destroyworkspacev2":
		id, _ := strconv.Atoi(e.Data[:strings.IndexRune(e.Data, ',')])
		_, ok := wsMod.ws[id]
		return ok

	case "renameworkspace":
		id, _ := strconv.Atoi(e.Data[:strings.IndexRune(e.Data, ',')])
		_, ok := wsMod.ws[id]
		return ok
	}

	return true
}

func (wsMod *HyprWorkspaceModule) setActiveWorkspace(id int) {
	wsMod.ws[wsMod.activeId].active = false
	wsMod.ws[id].active = true
	wsMod.ws[id].urgent = false
	wsMod.activeId = id
}

func (wsMod *HyprWorkspaceModule) createWorkspace(id int, name string) {
	wsMod.ws[id] = &WorkspaceState{id, name, false, false}
	if name == "special:magic" {
		wsMod.specialId = id
		wsMod.special = true
	}

}
func (wsMod *HyprWorkspaceModule) destroyWorkspace(id int) {
	delete(wsMod.ws, id)
	if id == wsMod.specialId {
		wsMod.special = false
	}
}

func (wsMod *HyprWorkspaceModule) isSpecialWorkspaceActive() bool {
	if wsMod.special {
		return wsMod.ws[wsMod.specialId].active
	}
	return false
}

func (wsMod *HyprWorkspaceModule) activateSpecialWorkspace(name string) {
	if name == "" {
		wsMod.ws[wsMod.specialId].active = false
	} else {
		wsMod.ws[wsMod.specialId].active = true
	}
}

func (wsMod *HyprWorkspaceModule) setWorkspaceUrgent(address string) {
	clients := hypr.GetClients()
	for _, client := range clients {
		client_address, _ := strings.CutPrefix(client.Address, "0x")
		if client_address == address {
			wsMod.ws[client.Workspace.Id].urgent = true
		}
	}
}

func (wsMod *HyprWorkspaceModule) handleHyprEvent(e hypr.HyprEvent) bool {
	switch e.Event {
	case "workspacev2":
		id, _ := strconv.Atoi(e.Data[:strings.IndexRune(e.Data, ',')])
		wsMod.setActiveWorkspace(id)
	case "createworkspacev2":
		id_str, name, _ := strings.Cut(e.Data, ",")
		id, _ := strconv.Atoi(id_str)
		wsMod.createWorkspace(id, name)
	case "destroyworkspacev2":
		id, _ := strconv.Atoi(e.Data[:strings.IndexRune(e.Data, ',')])
		wsMod.destroyWorkspace(id)
	case "activespecial":
		wsMod.activateSpecialWorkspace(e.Data[:strings.IndexRune(e.Data, ',')])
	case "urgent":
		wsMod.setWorkspaceUrgent(e.Data)
	default:
		return false
	}
	return true
}

func (wsMod *HyprWorkspaceModule) Render() []modules.EventCell {
	var wss []int
	for k := range wsMod.ws {
		if k > 0 {
			wss = append(wss, k)
		}
	}

	slices.Sort(wss)

	var r []modules.EventCell

	if wsMod.isSpecialWorkspaceActive() {
		var t1 modules.EventCell
		var t2 modules.EventCell
		var t3 modules.EventCell
		t1.C = ' '
		t2.C = 'S'
		t3.C = ' '

		t1.Style = modules.SPECIAL
		t2.Style = modules.SPECIAL
		t3.Style = modules.SPECIAL

		t1.Mod = wsMod
		t2.Mod = wsMod
		t3.Mod = wsMod
		r = append(r, t1, t2, t3)
	}

	for _, id := range wss {
		var t1 modules.EventCell
		var t2 modules.EventCell
		var t3 modules.EventCell
		t1.C = ' '
		t2.C = rune(wsMod.ws[id].name[0])
		t3.C = ' '
		if wsMod.ws[id].active {
			t1.Style = modules.ACTIVE
			t2.Style = modules.ACTIVE
			t3.Style = modules.ACTIVE
		}
		if wsMod.ws[id].urgent {
			t1.Style = modules.URGENT
			t2.Style = modules.URGENT
			t3.Style = modules.URGENT
		}

		t1.Metadata = wsMod.ws[id].name
		t2.Metadata = wsMod.ws[id].name
		t3.Metadata = wsMod.ws[id].name
		t1.Mod = wsMod
		t2.Mod = wsMod
		t3.Mod = wsMod
		r = append(r, t1, t2, t3)
	}

	return r
}
