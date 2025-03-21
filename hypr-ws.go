package main

import (
	"slices"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

var (
	URGENT  = tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorBlack)
	ACTIVE  = tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
	SPECIAL = tcell.StyleDefault.Background(tcell.ColorDarkGreen).Foreground(tcell.ColorWhite)
	DEFAULT = tcell.StyleDefault
)

type WSModuleWS struct {
	id     int
	name   string
	active bool
	urgent bool
}

type WSModule struct {
	ws        map[int]*WSModuleWS
	activeId  int
	specialId int
	special   bool
}

func (wsm *WSModule) RefreshWS() {
	wsm.ws = make(map[int]*WSModuleWS)

	workspaces := GetWorkspaces()
	active := GetActiveWorkspace()

	for _, ws := range workspaces {
		wsm.ws[ws.Id] = &WSModuleWS{ws.Id, ws.Name, ws.Id == active.Id, false}
		if ws.Name == "special:magic" {
			wsm.specialId = ws.Id
			wsm.special = true
		}
		if ws.Id == active.Id {
			wsm.activeId = ws.Id
		}
	}
}

func (wsm *WSModule) ValidateEvent(e HyprEvent) bool {
	switch e.event {
	case "workspacev2":
		id, _ := strconv.Atoi(e.data[:strings.IndexRune(e.data, ',')])
		_, ok := wsm.ws[id]
		return ok

	case "focusedmonv2":
		id, _ := strconv.Atoi(e.data[strings.LastIndex(e.data, ",")+1:])
		_, ok := wsm.ws[id]
		return ok

	case "createworkspacev2":
		id, _ := strconv.Atoi(e.data[:strings.IndexRune(e.data, ',')])
		_, ok := wsm.ws[id]
		return !ok

	case "destroyworkspacev2":
		id, _ := strconv.Atoi(e.data[:strings.IndexRune(e.data, ',')])
		_, ok := wsm.ws[id]
		return ok

	case "renameworkspace":
		id, _ := strconv.Atoi(e.data[:strings.IndexRune(e.data, ',')])
		_, ok := wsm.ws[id]
		return ok
	}

	return true
}

func (wsm *WSModule) SetActive(id int) {
	wsm.ws[wsm.activeId].active = false
	wsm.ws[id].active = true
	wsm.ws[id].urgent = false
	wsm.activeId = id
}

func (wsm *WSModule) CreateWS(id int, name string) {
	wsm.ws[id] = &WSModuleWS{id, name, false, false}
	if name == "special:magic" {
		wsm.specialId = id
		wsm.special = true
	}

}
func (wsm *WSModule) DestroyWS(id int) {
	delete(wsm.ws, id)
	if id == wsm.specialId {
		wsm.special = false
	}
}

func (wsm *WSModule) IsSpecialActive() bool {
	if wsm.special {
		return wsm.ws[wsm.specialId].active
	}
	return false
}

func (wsm *WSModule) Special(name string) {
	if name == "" {
		wsm.ws[wsm.specialId].active = false
	} else {
		wsm.ws[wsm.specialId].active = true
	}
}

func (wsm *WSModule) Urgent(address string) {
	clients := GetClients()
	for _, client := range clients {
		client_address, _ := strings.CutPrefix(client.Address, "0x")
		if client_address == address {
			wsm.ws[client.Workspace.Id].urgent = true
		}
	}
}

func (wsm *WSModule) HandleEvent(e HyprEvent) bool {
	switch e.event {
	case "workspacev2":
		id, _ := strconv.Atoi(e.data[:strings.IndexRune(e.data, ',')])
		wsm.SetActive(id)
		return true
	case "createworkspacev2":
		id_str, name, _ := strings.Cut(e.data, ",")
		id, _ := strconv.Atoi(id_str)
		wsm.CreateWS(id, name)
		return true
	case "destroyworkspacev2":
		id, _ := strconv.Atoi(e.data[:strings.IndexRune(e.data, ',')])
		wsm.DestroyWS(id)
		return true
	case "activespecial":
		wsm.Special(e.data[:strings.IndexRune(e.data, ',')])
		return true
	case "urgent":
		wsm.Urgent(e.data)
		return true
	}

	return false
}

func (wsm *WSModule) Render() []*EventCell {
	var wss []int
	for k := range wsm.ws {
		if k > 0 {
			wss = append(wss, k)
		}
	}

	slices.Sort(wss)

	var r []*EventCell

	if wsm.IsSpecialActive() {
		var t1 EventCell
		var t2 EventCell
		var t3 EventCell
		t1.c = ' '
		t2.c = 'S'
		t3.c = ' '

		t1.style = SPECIAL
		t2.style = SPECIAL
		t3.style = SPECIAL

		r = append(r, &t1, &t2, &t3)
	}

	for _, id := range wss {
		var t1 EventCell
		var t2 EventCell
		var t3 EventCell
		t1.c = ' '
		t2.c = rune(wsm.ws[id].name[0])
		t3.c = ' '
		if wsm.ws[id].active {
			t1.style = ACTIVE
			t2.style = ACTIVE
			t3.style = ACTIVE
		}
		if wsm.ws[id].urgent {
			t1.style = URGENT
			t2.style = URGENT
			t3.style = URGENT
		}

		t1.metadata = wsm.ws[id].name
		t2.metadata = wsm.ws[id].name
		t3.metadata = wsm.ws[id].name
		r = append(r, &t1, &t2, &t3)
	}

	return r
}
