package tui

import (
	"github.com/codelif/pawbar/internal/modules"
	"github.com/gdamore/tcell/v2"
)

func RenderBar(scr tcell.Screen, l, r []modules.Module, cells []modules.EventCell) {
	w, _ := scr.Size()
	for i := range w {
		cells[i] = modules.EventCell{C: ' ', Style: modules.DEFAULT, Metadata: "", Mod: nil}
		scr.SetContent(i, 0, ' ', nil, modules.DEFAULT)
	}

	p := 0
	for _, mod := range r {
		mod_render := mod.Render()
		len_mod := len(mod_render)
		for i := range len_mod {
			if p < w {
				c := mod_render[len_mod-i-1]
				scr.SetContent(w-p-1, 0, c.C, nil, c.Style)
				cells[w-p-1] = c
				p++
			}
		}
	}

	h := 0
	available := w - p
	ellipsized := false

	for _, mod := range l {
		for _, c := range mod.Render() {
			if h < available-3 {
				scr.SetContent(h, 0, c.C, nil, c.Style)
				cells[h] = c
				h++
			} else {
				if !ellipsized && h < available {
					for i := 0; i < 3 && h < available; i++ {
						scr.SetContent(h, 0, '.', nil, c.Style)
						cells[h] = modules.EventCell{C: '.', Style: c.Style}
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
