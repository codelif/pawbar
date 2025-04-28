package config

import (
	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/modules/backlight"
	"github.com/codelif/pawbar/internal/modules/battery"
	"github.com/codelif/pawbar/internal/modules/clock"
	"github.com/codelif/pawbar/internal/modules/cpu"
	"github.com/codelif/pawbar/internal/modules/disk"
	// "github.com/codelif/pawbar/internal/modules/hyprtitle"
	// "github.com/codelif/pawbar/internal/modules/hyprws"
	// "github.com/codelif/pawbar/internal/modules/locale"
	"github.com/codelif/pawbar/internal/modules/ram"
	// "github.com/codelif/pawbar/internal/modules/i3ws"
	// "github.com/codelif/pawbar/internal/modules/i3title"
)

var moduleFactories = map[string]func() modules.Module{
	"clock": clock.New,
	// "hyprtitle": hyprtitle.New,
	// "hyprws":    hyprws.New,
	"battery":   battery.New,
	"backlight": backlight.New,
	"ram":       ram.New,
	"cpu":       cpu.New,
	// "locale":    locale.New,
	"disk":      disk.New,
	// "i3ws":      i3ws.New,
	// "i3title":   i3title.New,
	"space": func() modules.Module {
		return modules.NewStaticModule(
			"space",
			[]modules.EventCell{
				{C: modules.ECSPACE.C},
			},
			nil,
		)
	},
	"sep": func() modules.Module {
		return modules.NewStaticModule(
			"sep",
			[]modules.EventCell{
				{C: modules.ECSPACE.C},
				{C: vaxis.Cell{
					Character: vaxis.Character{
						Grapheme: "│",
						Width:    1},
				}},
				{C: modules.ECSPACE.C},
			}, nil,
		)
	},
}
