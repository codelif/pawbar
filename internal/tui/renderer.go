package tui

import (
	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/modules"
)

var (
	modMap        = make(map[modules.Module][]modules.EventCell)
	state         []modules.EventCell
	leftModules   []modules.Module
	middleModules []modules.Module
	rightModules  []modules.Module
	width, height int
)

func State() []modules.EventCell {
	return state
}

// can be called again
func Init(w, h int, l, m, r []modules.Module) {
	width = w
	height = h

	leftModules = l
	middleModules = m
	rightModules = r

	state = make([]modules.EventCell, width)
	refreshModMap(leftModules, middleModules, rightModules)
}

func Resize(w, h int) {
	width = w
	height = h

	state = make([]modules.EventCell, width)
}

func FullRender(win vaxis.Window) {
	refreshModMap(leftModules, middleModules, rightModules)
	render(win)
}

func PartialRender(win vaxis.Window, m modules.Module) {
	modMap[m] = m.Render()
	render(win)
}

func refreshModMap(l, m, r []modules.Module) {
	for _, mod := range append(append(l, m...), r...) {
		modMap[mod] = mod.Render()
	}
}

func render(win vaxis.Window) {
	for i := range width {
		state[i] = modules.ECSPACE
	}
	win.Clear()

	leftCells := flatten(leftModules)
	midCells := flatten(middleModules)
	rightCells := flatten(rightModules)

	leftW := totalWidth(leftCells)
	midW := totalWidth(midCells)
	rightW := totalWidth(rightCells)

	if rightW > width {
		rightCells = trimStart(rightCells, width)
		rightW = totalWidth(rightCells)
		midCells = nil
	}

	if leftW > width-rightW {
		leftCells = trimEnd(leftCells, width-rightW)
		leftW = totalWidth(leftCells)
		midCells = nil
	}

	if leftW+rightW == width {
		midCells = nil
	}

	leftStart := 0
	rightStart := width - rightW

	midStart := (width - midW) / 2
	midEnd := midStart + midW

	if len(midCells) > 0 && leftW >= midStart {
		ell := modules.Cell('…', vaxis.Style{})
		midCells = append([]modules.EventCell{{C: ell}}, midCells[leftW-midStart+1:]...)
		midW = totalWidth(midCells)
		midStart = leftW
	}

	if len(midCells) > 0 && midEnd >= rightStart {
		ell := modules.Cell('…', vaxis.Style{})
		midCells = append(midCells[:midW-midEnd+rightStart-1], modules.EventCell{C: ell})
	}

	x := midStart
	for _, c := range midCells {
		x = writeCell(win, x, c)
	}

	x = leftStart
	for _, c := range leftCells {
		x = writeCell(win, x, c)
	}

	x = rightStart
	for _, c := range rightCells {
		x = writeCell(win, x, c)
	}
}

func writeCell(win vaxis.Window, x int, c modules.EventCell) int {
	win.SetCell(x, 0, c.C)
	state[x] = c

	for w := 1; w < c.C.Width; w++ {
		empty := vaxis.Cell{Style: c.C.Style}
		win.SetCell(x+w, 0, empty)
		state[x+w] = modules.EventCell{
			C:          empty,
			Metadata:   c.Metadata,
			Mod:        c.Mod,
			MouseShape: c.MouseShape,
		}
	}
	return x + c.C.Width
}

func flatten(mods []modules.Module) []modules.EventCell {
	// each module will probably require more than 3 cells
	// ws with 1 workspace requires 3, so most of them will
	// take more than that right? right? (foreshadowing)
	out := make([]modules.EventCell, 0, len(mods)*3)
	for _, m := range mods {
		out = append(out, modMap[m]...)
	}
	return out
}

// this keeps account for grapheme widths
// so this is safe for anchor calculations; probably?
func totalWidth(cells []modules.EventCell) int {
	w := 0
	for _, c := range cells {
		w += c.C.Width
	}
	return w
}

// also adds ellipsis at the start. ellipsis huh, weird word
func trimStart(cells []modules.EventCell, limit int) []modules.EventCell {
	w := 0
	for i := len(cells) - 1; i >= 0; i-- {
		c := cells[i]
		if w+c.C.Width > limit {
			ell := modules.Cell('…', vaxis.Style{})
			return append([]modules.EventCell{{C: ell}}, cells[i+1:]...)
		}
		w += c.C.Width
	}
	return cells
}

// also adds ellipsis at the end. ellipsis huh, indeed a weird word
func trimEnd(cells []modules.EventCell, limit int) []modules.EventCell {
	w := 0
	for i, c := range cells {
		if w+c.C.Width > limit {
			ell := modules.Cell('…', vaxis.Style{})
			return append(cells[:i-1], modules.EventCell{C: ell})
		}
		w += c.C.Width
	}
	return cells
}
