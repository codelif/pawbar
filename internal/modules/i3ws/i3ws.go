package i3ws

import (
	"slices"

	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/services/i3"
	"github.com/gdamore/tcell/v2"
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
	receive   chan bool
	send      chan modules.Event
	ievent    chan i3.I3Event
	ws        map[int]*WorkspaceState
	activeId  int
	specialId int
	special   bool
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

	wsMod.receive = make(chan bool)
	wsMod.send = make(chan modules.Event)
	wsMod.ievent = make(chan i3.I3Event)

	wsMod.refreshWorkspaceCache()
	go func() {
		for {
			select {
			case e := <-wsMod.send:
				switch ev := e.TcellEvent.(type) {
				case *tcell.EventMouse:
					btns := ev.Buttons()
					if btns == tcell.Button1 {
						go i3.GoToWorkspace(e.Cell.Metadata)
					}
				}
			case <-wsMod.ievent:
					wsMod.receive <- true			
			}
		}
	}()
	return wsMod.receive, wsMod.send, nil
}

func (wsMod *i3WorkspaceModule) refreshWorkspaceCache() {
	wsMod.ws = make(map[int]*WorkspaceState)

	workspaces := i3.GetWorkspaces()
	active := i3.GetActiveWorkspace()

	for _, ws := range workspaces {
		wsMod.ws[ws.Id] = &WorkspaceState{ws.Id, ws.Name, ws.Id == active.Id, false}

		// add special/scratchpad if any

		if ws.Id == active.Id {
			wsMod.activeId = ws.Id
		}
	}
}

func (wsMod *i3WorkspaceModule) Render() []modules.EventCell {
	var wss []int
	for k := range wsMod.ws {
		if k > 0 {
			wss = append(wss, k)
		}
	}

	slices.Sort(wss)

	var r []modules.EventCell

	for _, id := range wss {
		var t1 modules.EventCell
		var t2 modules.EventCell
		var t3 modules.EventCell
		t1.C = ' '
		t2.C = rune(wsMod.ws[id].name[0])
		t3.C = ' '
		if wsMod.ws[id].active {
			t1.Style = modules.ACTIVE.Reverse(true)
			t2.Style = modules.ACTIVE.Reverse(true)
			t3.Style = modules.ACTIVE.Reverse(true)
		}
		if wsMod.ws[id].urgent && !wsMod.ws[id].active {
			t1.Style = modules.URGENT.Reverse(true)
			t2.Style = modules.URGENT.Reverse(true)
			t3.Style = modules.URGENT.Reverse(true)
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

