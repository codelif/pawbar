// Copyright (c) 2025 Nekorg All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// SPDX-License-Identifier: bsd

package tui

import (
	"git.sr.ht/~rockorager/vaxis"
	"github.com/nekorg/pawbar/internal/config"
	"github.com/nekorg/pawbar/internal/modules"
)

var (
	modMap        = make(map[modules.Module][]modules.EventCell)
	state         []modules.EventCell
	leftModules   []modules.Module
	middleModules []modules.Module
	rightModules  []modules.Module
	width, height int

	truncOrder    []string
	useEllipsis   bool
	ellipsisCells []modules.EventCell
	ellipsisWidth int
)

type anchor int

const (
	left anchor = iota
	middle
	right
)

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
	useEllipsis = barCfg.EnableEllipsis == nil || *barCfg.EnableEllipsis
	ellipsisCells = stringToEC(barCfg.Ellipsis)
	ellipsisWidth = totalWidth(ellipsisCells)

	state = make([]modules.EventCell, width+1) // sometimes kitty can report mouse events outside reported width, like at the edge, idk why.
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
				ellW = ellipsisWidth
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
				if space-2*ellW > 0 {
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
					visible = append(clone(ellipsisCells), visible...)
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
					visible = append(visible, ellipsisCells...)
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
