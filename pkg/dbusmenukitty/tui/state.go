package tui

import (
	"time"

	"github.com/nekorg/pawbar/pkg/dbusmenukitty/menu"
)

type MenuState struct {
	items          []menu.Item
	mouseX         int
	mouseY         int
	mousePixelX    int
	mousePixelY    int
	lastMouseY     int
	mousePressed   bool
	mouseOnSurface bool
	hoverTimer     *time.Timer
	hoverItemId    int32
	ppc            menu.PPC
	size           menu.Size
}

func NewMenuState() *MenuState {
	return &MenuState{
		mouseX:      -1,
		mouseY:      -1,
		mousePixelX: -1,
		mousePixelY: -1,
		lastMouseY:  -1,
		hoverItemId: 0,
	}
}

func (m *MenuState) cancelHoverTimer() {
	if m.hoverTimer != nil {
		m.hoverTimer.Stop()
		m.hoverTimer = nil
	}
	// Don't clear hoverItemId here - it's needed to track active submenus
}

func (m *MenuState) isValidItemIndex(index int) bool {
	return index >= 0 && index < len(m.items)
}

func (m *MenuState) isSelectableItem(index int) bool {
	return m.isValidItemIndex(index) &&
		m.items[index].Type != menu.ItemSeparator &&
		m.items[index].Enabled ||
		!m.mouseOnSurface
}

func (m *MenuState) getCurrentItem() *menu.Item {
	if !m.isValidItemIndex(m.mouseY) {
		return nil
	}
	return &m.items[m.mouseY]
}

func (m *MenuState) navigateUp() {
	// Cancel any pending hover when navigating
	m.cancelHoverTimer()

	if m.mouseY > 0 {
		m.mouseY--
		// Skip separators
		for m.mouseY > 0 && m.items[m.mouseY].Type == menu.ItemSeparator {
			m.mouseY--
		}
	}
}

func (m *MenuState) navigateDown() {
	// Cancel any pending hover when navigating
	m.cancelHoverTimer()

	if m.mouseY < len(m.items)-1 {
		m.mouseY++
		// Skip separators
		for m.mouseY < len(m.items)-1 && m.items[m.mouseY].Type == menu.ItemSeparator {
			m.mouseY++
		}
	}
}
