package tui

import (
	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/modules"
)

var SPACE = modules.EventCell{
	C: vaxis.Cell{Character: vaxis.Character{
		Grapheme: " ",
		Width:    1,
	}},
	Metadata: "",
	Mod:      nil,
}
var DOT = modules.EventCell{
	C: vaxis.Cell{Character: vaxis.Character{
		Grapheme: ".",
		Width:    1,
	}},
	Metadata: "",
	Mod:      nil,
}

var modMap = make(map[string][]modules.EventCell)
var state []modules.EventCell
var leftModules []modules.Module
var rightModules []modules.Module
var width, height int

func State() []modules.EventCell {
	return state
}

// Can be recalled
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
		state[i] = SPACE
		win.SetCell(i, 0, SPACE.C)
	}

	p := 0
	for _, mod := range rightModules {

		mod_render := modMap[mod.Name()]
		len_mod := len(mod_render)
		for i := range len_mod {
			if p < width {
				c := mod_render[len_mod-i-1]
				win.SetCell(width-p-1, 0, c.C)
				state[width-p-1] = c
				p++
			}
		}
	}

	h := 0
	available := width - p
	ellipsized := false

	for _, mod := range leftModules {
		for _, c := range modMap[mod.Name()] {
			if h < available-3 {
				win.SetCell(h, 0, c.C)
				state[h] = c
				h++
			} else {
				if !ellipsized && h < available {
					for i := 0; i < 3 && h < available; i++ {
						state[h] = DOT
						win.SetCell(h, 0, DOT.C)
						h++
					}
					ellipsized = true
				}
				break
			}
		}
		if ellipsized {
			break
		}
	}
}
