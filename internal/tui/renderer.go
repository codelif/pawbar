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

// func refreshModMap(l, r []modules.Module){
//   for _, m := range append(l, r...) {
//     modMap[m.Name()]
//   }

// }


func RenderBar(scr vaxis.Window, l, r []modules.Module, m modules.Module, cells []modules.EventCell) {
	w, _ := scr.Size()
	for i := range w {
		cells[i] = SPACE
		scr.SetCell(i, 0, SPACE.C)
	}
  
	p := 0
	for _, mod := range r {
    if m != nil &&  {

    }else {
      mod_render := mod.Render()
    }
		len_mod := len(mod_render)
		for i := range len_mod {
			if p < w {
				c := mod_render[len_mod-i-1]
				scr.SetCell(w-p-1, 0, c.C)
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
				scr.SetCell(h, 0, c.C)
				cells[h] = c
				h++
			} else {
				if !ellipsized && h < available {
					for i := 0; i < 3 && h < available; i++ {
						cells[h] = DOT
						scr.SetCell(h, 0, DOT.C)
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
