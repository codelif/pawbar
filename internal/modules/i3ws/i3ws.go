package i3ws

import (
	"errors"
	"fmt"
	"slices"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/services/i3"
)

type WorkspaceState struct {
	id     int
	name   string
	active bool
	urgent bool
}

func New() modules.Module {
	return &i3WorkspaceModule{}
}

type i3WorkspaceModule struct {
	receive  chan bool
	send     chan modules.Event
	ievent   chan interface{}
	ws       map[int]*WorkspaceState
	activeId int
}

func (wsMod *i3WorkspaceModule) Name() string {
	return "i3ws"
}

func (wsMod *i3WorkspaceModule) Dependencies() []string {
	return []string{"i3"}
}

func (wsMod *i3WorkspaceModule) Channels() (<-chan bool, chan<- modules.Event) {
	return wsMod.receive, wsMod.send
}

func (wsMod *i3WorkspaceModule) Run() (<-chan bool, chan<- modules.Event, error) {

	service, ok := i3.GetService()
	if !ok {
		return nil, nil, errors.New("i3 service not available")
	}

	wsMod.receive = make(chan bool)
	wsMod.send = make(chan modules.Event)
	wsMod.ievent = make(chan interface{})

	service.RegisterChannel("workspaces", wsMod.ievent)

	wsMod.refreshWorkspaceCache()

	go func() {
		for {
			select {
			case e := <-wsMod.send:
				switch ev := e.VaxisEvent.(type) {
				case vaxis.Mouse:
					wsMod.handleMouseEvent(e, ev)
				}

			case raw := <-wsMod.ievent:
				switch evt := raw.(type) {
				case i3.I3Event:
					fmt.Println("event of type", evt)
					wsMod.refreshWorkspaceCache()
					wsMod.receive <- true
				default:
					fmt.Println("Unknown event type received on workspace ievent:", raw)
				}
			}
		}
	}()

	return wsMod.receive, wsMod.send, nil
}

func (wsMod *i3WorkspaceModule) handleMouseEvent(e modules.Event, ev vaxis.Mouse) {
	if ev.EventType != vaxis.EventPress {
		return
	}

	switch ev.Button {
	case vaxis.MouseLeftButton:
		go i3.GoToWorkspace(e.Cell.Metadata)
	}

}

func (wsMod *i3WorkspaceModule) refreshWorkspaceCache() {
	wsMod.ws = make(map[int]*WorkspaceState)

	workspaces := i3.GetWorkspaces()
	active := i3.GetActiveWorkspace()

	for _, ws := range workspaces {
		wsMod.ws[ws.Id] = &WorkspaceState{ws.Id, ws.Name, ws.Id == active.Id, ws.Urgent}

		// add special/scratchpad if any

		if ws.Id == active.Id {
			wsMod.activeId = ws.Id
		}
	}
}

var SPECIAL = vaxis.Style{Foreground: modules.ACTIVE, Background: modules.SPECIAL}
var ACTIVE = vaxis.Style{Foreground: modules.BLACK, Background: modules.ACTIVE}
var URGENT = vaxis.Style{Foreground: modules.BLACK, Background: modules.URGENT}

func (wsMod *i3WorkspaceModule) Render() []modules.EventCell {
	var wss []int
	for k := range wsMod.ws {
		wss = append(wss, k)
	}

	slices.Sort(wss)

	var r []modules.EventCell

	for _, id := range wss {
		var t1 modules.EventCell
		var t2 modules.EventCell
		var t3 modules.EventCell
		t1.C = modules.Cell(' ', vaxis.Style{})
		t2.C = modules.Cell(rune(wsMod.ws[id].name[0]), vaxis.Style{})
		t3.C = modules.Cell(' ', vaxis.Style{})
		if wsMod.ws[id].active {
			t1.C.Style = ACTIVE
			t2.C.Style = ACTIVE
			t3.C.Style = ACTIVE
		}
		if wsMod.ws[id].urgent && !wsMod.ws[id].active {
			t1.C.Style = URGENT
			t2.C.Style = URGENT
			t3.C.Style = URGENT
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
