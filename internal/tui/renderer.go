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
	for _, mod := range l {
		for _, c := range mod.Render() {
			// utils.Logger.Printf("%s: [%d]: '%c', '%s'\n", mod.Name(), p, c.c, c.m.Name())
			if p < w {
				scr.SetContent(p, 0, c.C, nil, c.Style)
				cells[p] = c
				p++
			}
		}
	}

	p = 0
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
}
