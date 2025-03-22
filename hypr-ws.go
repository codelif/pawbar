package main

import (
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type WS struct {
	id     int
	name   string
	active bool
	urgent bool
}

func NewHyprWorkspaces() Module {
	return &HyprWorkspaces{}
}

type HyprWorkspaces struct {
	hypr      *HyprService
	receive   chan bool
	send      chan Event
	hevent    chan HyprEvent
	ws        map[int]*WS
	activeId  int
	specialId int
	special   bool
}

func (hyprws *HyprWorkspaces) Name() string {
	return "hypr-ws"
}

func (hypews *HyprWorkspaces) Dependencies() []string {
	return []string{"hypr"}
}

func (hyprws *HyprWorkspaces) Run() (<-chan bool, chan<- Event, error) {
	service, ok := serviceRegistry["hypr"].(*HyprService)
	if !ok {
		return nil, nil, errors.New("Hypr service not available")
	}
  hyprws.hypr = service

	hyprws.receive = make(chan bool)
	hyprws.send = make(chan Event)
	hyprws.hevent = make(chan HyprEvent)
	for _, e := range []string{"workspacev2", "focusedmonv2", "createworkspacev2", "destroyworkspacev2", "activespecial", "renameworkspace", "urgent"} {
		hyprws.hypr.RegisterChannel(e, hyprws.hevent)
	}
	hyprws.RefreshWS()
	go func() {
		for {
			select {
			case e := <-hyprws.send:
				switch ev := e.e.(type) {
				case *tcell.EventMouse:
					btns := ev.Buttons()
					if btns == tcell.Button1 {
						go GoToWorkspace(e.c.metadata)
					}
				}
			case h := <-hyprws.hevent:
				if !hyprws.ValidateEvent(h) {
					hyprws.RefreshWS()
				}
				if hyprws.HandleEvent(h) {
					hyprws.receive <- true
				}
			}
		}
	}()

	return hyprws.receive, hyprws.send, nil
}

func (hyprws *HyprWorkspaces) Channels() (<-chan bool, chan<- Event) {
	return hyprws.receive, hyprws.send
}

func (hyprws *HyprWorkspaces) RefreshWS() {
	hyprws.ws = make(map[int]*WS)

	workspaces := GetWorkspaces()
	active := GetActiveWorkspace()

	for _, ws := range workspaces {
		hyprws.ws[ws.Id] = &WS{ws.Id, ws.Name, ws.Id == active.Id, false}
		if ws.Name == "special:magic" {
			hyprws.specialId = ws.Id
			hyprws.special = true
		}
		if ws.Id == active.Id {
			hyprws.activeId = ws.Id
		}
	}
}

func (hyprws *HyprWorkspaces) ValidateEvent(e HyprEvent) bool {
	switch e.event {
	case "workspacev2":
		id, _ := strconv.Atoi(e.data[:strings.IndexRune(e.data, ',')])
		_, ok := hyprws.ws[id]
		return ok

	case "focusedmonv2":
		id, _ := strconv.Atoi(e.data[strings.LastIndex(e.data, ",")+1:])
		_, ok := hyprws.ws[id]
		return ok

	case "createworkspacev2":
		id, _ := strconv.Atoi(e.data[:strings.IndexRune(e.data, ',')])
		_, ok := hyprws.ws[id]
		return !ok

	case "destroyworkspacev2":
		id, _ := strconv.Atoi(e.data[:strings.IndexRune(e.data, ',')])
		_, ok := hyprws.ws[id]
		return ok

	case "renameworkspace":
		id, _ := strconv.Atoi(e.data[:strings.IndexRune(e.data, ',')])
		_, ok := hyprws.ws[id]
		return ok
	}

	return true
}

func (hyprws *HyprWorkspaces) SetActive(id int) {
	hyprws.ws[hyprws.activeId].active = false
	hyprws.ws[id].active = true
	hyprws.ws[id].urgent = false
	hyprws.activeId = id
}

func (hyprws *HyprWorkspaces) CreateWS(id int, name string) {
	hyprws.ws[id] = &WS{id, name, false, false}
	if name == "special:magic" {
		hyprws.specialId = id
		hyprws.special = true
	}

}
func (hyprws *HyprWorkspaces) DestroyWS(id int) {
	delete(hyprws.ws, id)
	if id == hyprws.specialId {
		hyprws.special = false
	}
}

func (hyprws *HyprWorkspaces) IsSpecialActive() bool {
	if hyprws.special {
		return hyprws.ws[hyprws.specialId].active
	}
	return false
}

func (hyprws *HyprWorkspaces) Special(name string) {
	if name == "" {
		hyprws.ws[hyprws.specialId].active = false
	} else {
		hyprws.ws[hyprws.specialId].active = true
	}
}

func (hyprws *HyprWorkspaces) Urgent(address string) {
	clients := GetClients()
	for _, client := range clients {
		client_address, _ := strings.CutPrefix(client.Address, "0x")
		if client_address == address {
			hyprws.ws[client.Workspace.Id].urgent = true
		}
	}
}

func (hyprws *HyprWorkspaces) HandleEvent(e HyprEvent) bool {
	switch e.event {
	case "workspacev2":
		id, _ := strconv.Atoi(e.data[:strings.IndexRune(e.data, ',')])
		hyprws.SetActive(id)
	case "createworkspacev2":
		id_str, name, _ := strings.Cut(e.data, ",")
		id, _ := strconv.Atoi(id_str)
		hyprws.CreateWS(id, name)
	case "destroyworkspacev2":
		id, _ := strconv.Atoi(e.data[:strings.IndexRune(e.data, ',')])
		hyprws.DestroyWS(id)
	case "activespecial":
		hyprws.Special(e.data[:strings.IndexRune(e.data, ',')])
	case "urgent":
		hyprws.Urgent(e.data)
	default:
		return false
	}
	return true
}

func (hyprws *HyprWorkspaces) Render() []EventCell {
	var wss []int
	for k := range hyprws.ws {
		if k > 0 {
			wss = append(wss, k)
		}
	}

	slices.Sort(wss)

	var r []EventCell

	if hyprws.IsSpecialActive() {
		var t1 EventCell
		var t2 EventCell
		var t3 EventCell
		t1.c = ' '
		t2.c = 'S'
		t3.c = ' '

		t1.style = SPECIAL
		t2.style = SPECIAL
		t3.style = SPECIAL

		t1.m = hyprws
		t2.m = hyprws
		t3.m = hyprws
		r = append(r, t1, t2, t3)
	}

	for _, id := range wss {
		var t1 EventCell
		var t2 EventCell
		var t3 EventCell
		t1.c = ' '
		t2.c = rune(hyprws.ws[id].name[0])
		t3.c = ' '
		if hyprws.ws[id].active {
			t1.style = ACTIVE
			t2.style = ACTIVE
			t3.style = ACTIVE
		}
		if hyprws.ws[id].urgent {
			t1.style = URGENT
			t2.style = URGENT
			t3.style = URGENT
		}

		t1.metadata = hyprws.ws[id].name
		t2.metadata = hyprws.ws[id].name
		t3.metadata = hyprws.ws[id].name
		t1.m = hyprws
		t2.m = hyprws
		t3.m = hyprws
		r = append(r, t1, t2, t3)
	}

	return r
}
