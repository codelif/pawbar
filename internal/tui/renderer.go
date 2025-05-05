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

	rightModulesLength := 0
outerRight:
	for _, mod := range rightModules {
		modRender := modMap[mod]
		modLength := len(modRender)
		for i := range modLength {
			if rightModulesLength >= width {
				break outerRight
			}
			c := modRender[modLength-i-1]
			win.SetCell(width-rightModulesLength-1, 0, c.C)
			state[width-rightModulesLength-1] = c
			rightModulesLength++

			if c.C.Width > 1 {
				emptyCell := vaxis.Cell{Style: c.C.Style}
				for range c.C.Width - 1 {
					win.SetCell(width-rightModulesLength-1, 0, emptyCell)

					state[width-rightModulesLength-1] = modules.EventCell{
						C:          emptyCell,
						Metadata:   c.Metadata,
						Mod:        c.Mod,
						MouseShape: c.MouseShape,
					}

					rightModulesLength++
				}
			}
		}
	}

<<<<<<< Updated upstream
	anchor := width / 2
	rightEdge := width - rightModulesLength
	available := rightEdge - anchor

	middleLeftLength, middleRightLength := 0, 0

outerMiddle:
	for _, mod := range middleModules {
		modRender := modMap[mod]
		modLength := len(modRender)
		half := modLength / 2

		for i := 0; i < half; i++ {
			if middleLeftLength >= anchor-(half-i) {
				for ; middleLeftLength < anchor; middleLeftLength++ {
					x := anchor - middleLeftLength - 1
					ell := modules.Cell('…', vaxis.Style{})
					win.SetCell(x, 0, ell)
					state[x] = modules.EventCell{C: ell, Mod: nil}
				}
				break outerMiddle
			}
=======
middleLeftLength, middleRightLength := 0, 0

// First, flatten all middle modules into a single slice
modRender := make([]modules.EventCell, 0, len(middleModules)*10)
for _, mod := range middleModules {
    modRender = append(modRender, modMap[mod]...)
}

anchor := width / 2
rightEdge := width - rightModulesLength
available := rightEdge - anchor
half := len(modRender) / 2
>>>>>>> Stashed changes

leftCells := modRender[:half]
rightCells := modRender[half:]

<<<<<<< Updated upstream
			if c.C.Width > 1 {
				empty := vaxis.Cell{Style: c.C.Style}
				for j := 0; j < c.C.Width-1; j++ {
					if middleLeftLength >= anchor-(half-i) {
						for ; middleLeftLength < anchor; middleLeftLength++ {
							x = anchor - middleLeftLength - 1
							ell := modules.Cell('…', vaxis.Style{})
							win.SetCell(x, 0, ell)
							state[x] = modules.EventCell{C: ell, Mod: nil}
						}
						break outerMiddle
					}
					x = anchor - middleLeftLength - 1
					win.SetCell(x, 0, empty)
					state[x] = modules.EventCell{
						C:          empty,
						Metadata:   c.Metadata,
						Mod:        c.Mod,
						MouseShape: c.MouseShape,
					}
					middleLeftLength++
				}
			}
		}

		for i := half; i < len(modRender); i++ {
			c := modRender[i]
			w := c.C.Width

			if middleRightLength+w > available {
				// overflow: draw one ellipsis at the very end and stop
				pos := anchor + available - 1
				ell := modules.Cell('…', vaxis.Style{})
				win.SetCell(pos, 0, ell)
				state[pos] = modules.EventCell{C: ell, Mod: nil}
				break outerMiddle
			}

			x := anchor + middleRightLength
			win.SetCell(x, 0, c.C)
			state[x] = c
			middleRightLength++

			for j := 1; j < w; j++ {
				x = anchor + middleRightLength
				empty := vaxis.Cell{Style: c.C.Style}
				win.SetCell(x, 0, empty)
				state[x] = modules.EventCell{
					C:          empty,
					Metadata:   c.Metadata,
					Mod:        c.Mod,
					MouseShape: c.MouseShape,
				}
				middleRightLength++
			}
		}
	}
=======
for i := len(leftCells) - 1; i >= 0; i-- {
    c := leftCells[i]
    w := c.C.Width
    if anchor - middleLeftLength - w < 0 {
        for ; middleLeftLength < anchor; middleLeftLength++ {
            x := anchor - middleLeftLength - 1
            ell := modules.Cell('…', vaxis.Style{})
            win.SetCell(x, 0, ell)
            state[x] = modules.EventCell{C: ell, Mod: nil}
        }
        break
    }

    for j := 0; j < w; j++ {
        x := anchor - middleLeftLength - 1
        if j == 0 {
            win.SetCell(x, 0, c.C)
            state[x] = c
        } else {
            pad := vaxis.Cell{Style: c.C.Style}
            win.SetCell(x, 0, pad)
            state[x] = modules.EventCell{C: pad, Metadata: c.Metadata, Mod: c.Mod, MouseShape: c.MouseShape}
        }
        middleLeftLength++
    }
}


for _, c := range rightCells {
    w := c.C.Width
    // check overflow on right
    if middleRightLength + w > available {
        pos := anchor + available - 1
        ell := modules.Cell('…', vaxis.Style{})
        win.SetCell(pos, 0, ell)
        state[pos] = modules.EventCell{C: ell, Mod: nil}
        break
    }
    for j := 0; j < w; j++ {
        x := anchor + middleRightLength
        if j == 0 {
            win.SetCell(x, 0, c.C)
            state[x] = c
        } else {
            pad := vaxis.Cell{Style: c.C.Style}
            win.SetCell(x, 0, pad)
            state[x] = modules.EventCell{C: pad, Metadata: c.Metadata, Mod: c.Mod, MouseShape: c.MouseShape}
        }
        middleRightLength++
    }
}
>>>>>>> Stashed changes

	leftModulesLength := 0
	availableForleft := width/2 - middleLeftLength
outerLeft:
	for _, mod := range leftModules {
		for _, c := range modMap[mod] {

			if leftModulesLength >= availableForleft-1 {
				for range availableForleft - leftModulesLength {
					state[leftModulesLength] = modules.EventCell{C: modules.Cell('…', vaxis.Style{}), Mod: nil}
					win.SetCell(leftModulesLength, 0, state[leftModulesLength].C)
					leftModulesLength++
				}
				break outerLeft
			}

			win.SetCell(leftModulesLength, 0, c.C)
			state[leftModulesLength] = c
			leftModulesLength++

			if c.C.Width > 1 {
				emptyCell := vaxis.Cell{Style: c.C.Style}
				for range c.C.Width - 1 {
					win.SetCell(leftModulesLength, 0, emptyCell)

					state[leftModulesLength] = modules.EventCell{
						C:          emptyCell,
						Metadata:   c.Metadata,
						Mod:        c.Mod,
						MouseShape: c.MouseShape,
					}

					leftModulesLength++
				}
			}
		}
	}
}
