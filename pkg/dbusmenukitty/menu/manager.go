package menu

import (
	"sync"

	"github.com/codelif/katnip"
	"github.com/fxamacker/cbor/v2"
)

type MenuManager struct {
	panels    []*katnip.Panel
	positions []Position
	mutex     sync.RWMutex
}


var globalManager = &MenuManager{}

func GetManager() *MenuManager {
	return globalManager
}

func (sm *MenuManager) AddPanel(panel *katnip.Panel, x, y int) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.panels = append(sm.panels, panel)
	sm.positions = append(sm.positions, Position{X: x, Y: y})
}

func (sm *MenuManager) RemovePanel(panel *katnip.Panel) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	for i, p := range sm.panels {
		if p == panel {
			for j := i + 1; j < len(sm.panels); j++ {
				sm.panels[j].Stop()
			}

			sm.panels = sm.panels[:i]
			sm.positions = sm.positions[:i]
			break
		}
	}
}

func (sm *MenuManager) CloseAllSubmenus() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if len(sm.panels) > 1 {
		for i := 1; i < len(sm.panels); i++ {
			closeMsg := Message{
				Type:    MsgMenuClose,
				Payload: MessagePayload{},
			}
			enc := cbor.NewEncoder(sm.panels[i].Writer())
			enc.Encode(closeMsg)
			// sm.panels[i].Stop()
		}
		sm.panels = sm.panels[:1]
		sm.positions = sm.positions[:1]
	}
}

func (sm *MenuManager) GetNextPosition() (int, int) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if len(sm.positions) == 0 {
		return 0, 0
	}

	lastPos := sm.positions[len(sm.positions)-1]
	return lastPos.X + 200, lastPos.Y // Adjust spacing as needed
}

func (sm *MenuManager) HandlePanelExit(panel *katnip.Panel) {
	sm.RemovePanel(panel)
}
