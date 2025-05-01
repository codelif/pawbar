package tui

import (
	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/modules"
)

var modMap = make(map[string][]modules.EventCell)
var state []modules.EventCell
var leftModules []modules.Module
var rightModules []modules.Module
var width, height int

func State() []modules.EventCell {
	return state
}

// can be called again
func Init(w, h int, l, r []modules.Module) {
	width = w
	height = h

	leftModules = l
	rightModules = r

	state = make([]modules.EventCell, width)
	refreshModMap(leftModules, rightModules)
}

func Resize(w, h int) {
	width = w
	height = h

	state = make([]modules.EventCell, width)
}

func FullRender(win vaxis.Window) {
	refreshModMap(leftModules, rightModules)
	render(win)
}

func PartialRender(win vaxis.Window, m modules.Module) {
	modMap[m.Name()] = m.Render()
	render(win)
}

func refreshModMap(l, r []modules.Module) {
	for _, m := range append(l, r...) {
		modMap[m.Name()] = m.Render()
	}
}

func render(win vaxis.Window) {
	for i := range width {
		state[i] = modules.ECSPACE
		win.SetCell(i, 0, modules.ECSPACE.C)
	}

	rightModulesLength := 0
outerRight:
	for _, mod := range rightModules {
		modRender := modMap[mod.Name()]
		modLength := len(modRender)
		for i := range modLength {
			if rightModulesLength >= width {
				break outerRight
			}
			c := modRender[modLength-i-1]
			win.SetCell(width-rightModulesLength-1, 0, c.C)
			state[width-rightModulesLength-1] = c
			rightModulesLength++
		}
	}

	leftModulesLength := 0
	available := width - rightModulesLength
outerLeft:
	for _, mod := range leftModules {
		for _, c := range modMap[mod.Name()] {

			if leftModulesLength >= available-1 {
				for range available - leftModulesLength {
					state[leftModulesLength] = modules.EventCell{C: modules.Cell('…', vaxis.Style{}), Mod: nil}
					win.SetCell(leftModulesLength, 0, state[leftModulesLength].C)
					leftModulesLength++
				}
				break outerLeft
			}

			win.SetCell(leftModulesLength, 0, c.C)
			state[leftModulesLength] = c
			leftModulesLength++
		}
	}
}
