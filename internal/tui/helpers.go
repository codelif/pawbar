// Copyright (c) 2025 Nekorg All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// SPDX-License-Identifier: bsd

package tui

import (
	"git.sr.ht/~rockorager/vaxis"
	"github.com/nekorg/pawbar/internal/modules"
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

func stringToEC(s string) []modules.EventCell {
	ecs := make([]modules.EventCell, 0, len(s))
	for _, c := range vaxis.Characters(s) {
		ecs = append(ecs, modules.EventCell{C: vaxis.Cell{Character: c}})
	}

	return ecs
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
		if ellipsisWidth >= w {
			return nil
		}
		w -= ellipsisWidth
	}
	acc := 0
	end := 0
	for ; end < len(cells) && acc < w; end++ {
		acc += totalWidth(cells[end : end+1])
	}
	trim := cells[:end]
	if ellipsis {
		trim = append(trim, ellipsisCells...)
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
		if ellipsisWidth >= w {
			return nil
		}
		w -= ellipsisWidth
	}
	acc := 0
	start := len(cells)
	for start > 0 && acc < w {
		start--
		acc += totalWidth(cells[start : start+1])
	}
	trim := cells[start:]
	if ellipsis {
		trim = append(clone(ellipsisCells), trim...)
	}
	return trim
}

// also adds ellipsis at the end and start. huh weird, where have I seen thi-
func trimMiddle(cells []modules.EventCell, w int, ellipsis bool) []modules.EventCell {
	if w <= 0 || totalWidth(cells) <= w {
		return cells
	}
	if ellipsis {
		ellW := ellipsisWidth * 2
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
			append(clone(ellipsisCells), trimmed...),
			ellipsisCells...,
		)
	}
	return trimmed
}

func clone(src []modules.EventCell) []modules.EventCell {
	out := make([]modules.EventCell, len(src))
	copy(out, src)
	return out
}
