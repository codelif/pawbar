package config

import (
	"github.com/codelif/pawbar/internal/modules"
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
	"space": func() modules.Module {
		return modules.NewStaticModule("space", []modules.EventCell{{C: ' ', Style: modules.DEFAULT}}, nil)
	},
	"sep": func() modules.Module {
		return modules.NewStaticModule("sep", []modules.EventCell{{C: '│', Style: modules.DEFAULT}}, nil)
	},
}
