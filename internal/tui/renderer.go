package tui

import (
	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
)

var (
	modMap        = make(map[modules.Module][]modules.EventCell)
	state         []modules.EventCell
	leftModules   []modules.Module
	middleModules []modules.Module
	rightModules  []modules.Module
	width, height int

	truncOrder  []string
	useEllipsis bool
)

type anchor int

const (
	left anchor = iota
	middle
	right
)

func anchorOf(s string) anchor {
	switch s {
	case "left":
		return left
	case "middle":
		return middle
	default:
		return right
	}
}

type block struct {
	cells []modules.EventCell
	side  anchor
}

func State() []modules.EventCell {
	return state
}

// can be called again
func Init(w, h int, l, m, r []modules.Module, barCfg config.BarSettings) {
	width = w
	height = h

	leftModules = l
	middleModules = m
	rightModules = r

	truncOrder = barCfg.TruncatePriority
	useEllipsis = barCfg.Ellipsis == nil || *barCfg.Ellipsis

	state = make([]modules.EventCell, width+1)
	refreshModMap(leftModules, middleModules, rightModules)
}

func Resize(w, h int) {
	width = w
	height = h

	state = make([]modules.EventCell, width+1)
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

// flattens modules into blocks sorted by priority.
func buildBlocks() []block {
	cells := map[string][]modules.EventCell{
		"left":   flatten(leftModules),
		"middle": flatten(middleModules),
		"right":  flatten(rightModules),
	}

	blocks := make([]block, 0, 3)
	for _, name := range truncOrder {
		blocks = append(blocks, block{
			cells: cells[name],
			side:  anchorOf(name),
		})
	}
	return blocks
}

func render(win vaxis.Window) {
	for i := range width {
		state[i] = modules.ECSPACE
	}
	win.Clear()

	blocks := buildBlocks()
	occ := make([]bool, width)

	mark := func(x, w int) {
		for i := 0; i < w && x+i < width; i++ {
			occ[x+i] = true
		}
	}

	for _, block := range blocks {
		if len(block.cells) == 0 {
			continue
		}

		fullW := totalWidth(block.cells)
		switch block.side {
		case left:
			free := 0
			for free < width && !occ[free] {
				free++
			}
			visible := block.cells
			if fullW > free {
				visible = trimStart(block.cells, free, useEllipsis)
			}
			if len(visible) == 0 {
				break
			}
			x := 0
			for _, r := range visible {
				next := writeCell(win, x, r)
				mark(x, next-x)
				x = next
			}

		case middle:
			start := (width - fullW) / 2
			if start < 0 {
				start = 0
			}
			end := start + fullW

			firstOcc, lastOcc := -1, -1
			for i := start; i < end && i < width; i++ {
				if occ[i] {
					if firstOcc == -1 {
						firstOcc = i
					}
					lastOcc = i
				}
			}
			if firstOcc == -1 {
				x := start
				for _, r := range block.cells {
					next := writeCell(win, x, r)
					mark(x, next-x)
					x = next
				}
				break
			}

			var visible []modules.EventCell
			var drawAt int

			ellW := 0
			if useEllipsis {
				ellW = modules.ECELLIPSIS.C.Width
			}
			switch {
			case firstOcc == start && lastOcc == end-1:
				gapStart := 0
				for gapStart < width && occ[gapStart] {
					gapStart++
				}
				gapEnd := width - 1
				for gapEnd >= 0 && occ[gapEnd] {
					gapEnd--
				}
				gapLen := gapEnd - gapStart + 1
				if gapLen <= 0 {
					return
				}
				space := gapLen
				if useEllipsis {
					space -= 2 * ellW
				}
				if space > 0 {
					visible := trimMiddle(block.cells, space, useEllipsis)

					drawAt := gapStart + (gapLen-totalWidth(visible))/2
					x := drawAt
					for _, r := range visible {
						next := writeCell(win, x, r)
						mark(x, next-x)
						x = next
					}
				}
				break
			case firstOcc == start:
				free := end - lastOcc - 1
				space := free
				if useEllipsis {
					space -= ellW
				}
				if space <= 0 {
					break
				}
				visible = trimEnd(block.cells, space, false)
				if useEllipsis {
					visible = append([]modules.EventCell{modules.ECELLIPSIS}, visible...)
				}
				drawAt = end - totalWidth(visible)

			case lastOcc == end-1 || firstOcc > start:
				free := firstOcc - start
				space := free
				if useEllipsis {
					space -= ellW
				}
				if space <= 0 {
					break
				}
				visible = trimStart(block.cells, space, false)
				if useEllipsis {
					visible = append(visible, modules.ECELLIPSIS)
				}
				drawAt = start

			default:
				break
			}

			x := drawAt
			for _, r := range visible {
				next := writeCell(win, x, r)
				mark(x, next-x)
				x = next
			}
		case right:
			free := 0
			for i := width - 1; i >= 0 && !occ[i]; i-- {
				free++
			}
			visible := block.cells
			if fullW > free {
				visible = trimEnd(block.cells, free, useEllipsis)
			}
			if len(visible) == 0 {
				break
			}
			renderW := totalWidth(visible)
			start := width - renderW
			x := start
			for _, r := range visible {
				next := writeCell(win, x, r)
				mark(x, next-x)
				x = next
			}
		}
	}
}

// writes cell and adds padding for grapheme's with >1 width
// returns x + {grapheme width}
func writeCell(win vaxis.Window, x int, c modules.EventCell) int {
	if x+c.C.Width > width {
		return x + c.C.Width
	}
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
func trimStart(cells []modules.EventCell, w int, ellipsis bool) []modules.EventCell {
	if w <= 0 {
		return nil
	}
	if totalWidth(cells) <= w {
		return cells
	}
	if ellipsis {
		ew := modules.ECELLIPSIS.C.Width
		if ew >= w {
			return nil
		}
		w -= ew
	}
	acc := 0
	end := 0
	for ; end < len(cells) && acc < w; end++ {
		acc += totalWidth(cells[end : end+1])
	}
	trim := cells[:end]
	if ellipsis {
		trim = append(trim, modules.ECELLIPSIS)
	}
	return trim
}

// also adds ellipsis at the end. ellipsis huh, indeed a weird word
func trimEnd(cells []modules.EventCell, w int, ellipsis bool) []modules.EventCell {
	if w <= 0 {
		return nil
	}
	if totalWidth(cells) <= w {
		return cells
	}
	if ellipsis {
		ew := modules.ECELLIPSIS.C.Width
		if ew >= w {
			return nil
		}
		w -= ew
	}
	acc := 0
	start := len(cells)
	for start > 0 && acc < w {
		start--
		acc += totalWidth(cells[start : start+1])
	}
	trim := cells[start:]
	if ellipsis {
		trim = append([]modules.EventCell{modules.ECELLIPSIS}, trim...)
	}
	return trim
}

// also adds ellipsis at the end and start. huh weird, where have I seen thi-
func trimMiddle(cells []modules.EventCell, w int, ellipsis bool) []modules.EventCell {
	if w <= 0 || totalWidth(cells) <= w {
		return cells
	}
	ellW := 0
	if ellipsis {
		ellW = modules.ECELLIPSIS.C.Width * 2
		if ellW >= w {
			return nil
		}
		w -= ellW
	}

	left, right := 0, len(cells)-1
	cur := totalWidth(cells)
	for cur > w && left < right {
		cur -= cells[left].C.Width
		left++
		if cur > w && left < right {
			cur -= cells[right].C.Width
			right--
		}
	}
	trimmed := cells[left : right+1]
	if ellipsis {
		trimmed = append(
			append([]modules.EventCell{modules.ECELLIPSIS}, trimmed...),
			modules.ECELLIPSIS,
		)
	}
	return trimmed
}
