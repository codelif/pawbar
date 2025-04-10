package config

import (
	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/modules/backlight"
	"github.com/codelif/pawbar/internal/modules/locale"
	"github.com/codelif/pawbar/internal/modules/ram"
	"github.com/codelif/pawbar/internal/modules/cpu"
	"github.com/codelif/pawbar/internal/modules/battery"
	"github.com/codelif/pawbar/internal/modules/clock"
	"github.com/codelif/pawbar/internal/modules/hyprtitle"
	"github.com/codelif/pawbar/internal/modules/hyprws"
)

var moduleFactories = map[string]func() modules.Module{
	"clock":     clock.New,
	"hyprtitle": hyprtitle.New,
	"hyprws":    hyprws.New,
	"battery":   battery.New,
	"backlight": backlight.New,
	"ram":       ram.New,
	"cpu":       cpu.New,
	"locale":    locale.New,
	"space": func() modules.Module {
		return modules.NewStaticModule("space", []modules.EventCell{{C: ' ', Style: modules.DEFAULT}}, nil)
	},
	"sep": func() modules.Module {
		return modules.NewStaticModule("sep", []modules.EventCell{{C: ' ', Style: modules.DEFAULT}, {C: '│', Style: modules.DEFAULT}, {C: ' ', Style: modules.DEFAULT}}, nil)
	},
}
