package ws

import (
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/lookup/colors"
	"github.com/codelif/pawbar/internal/modules"
)

func init() {
	config.RegisterModule("ws", defaultOptions, func(o Options) (modules.Module, error) { return &Module{opts: o}, nil })
}

type ActiveOptions struct {
	Fg config.Color `yaml:"fg"`
	Bg config.Color `yaml:"bg"`
}

type UrgentOptions struct {
	Fg config.Color `yaml:"fg"`
	Bg config.Color `yaml:"bg"`
}

type SpecialOptions struct {
	Fg config.Color `yaml:"fg"`
	Bg config.Color `yaml:"bg"`
}

type Options struct {
	Fg      config.Color                      `yaml:"fg"`
	Bg      config.Color                      `yaml:"bg"`
	Cursor  config.Cursor                     `yaml:"cursor"`
	Format  config.Format                     `yaml:"format"`
	Special SpecialOptions                    `yaml:"special"`
	Active  ActiveOptions                     `yaml:"active"`
	Urgent  UrgentOptions                     `yaml:"urgent"`
	OnClick config.MouseActions[MouseOptions] `yaml:"onmouse"`
}

type MouseOptions struct {
	Fg     *config.Color  `yaml:"fg"`
	Bg     *config.Color  `yaml:"bg"`
	Cursor *config.Cursor `yaml:"cursor"`
	Format *config.Format `yaml:"format"`
}

func defaultOptions() Options {
	fw, _ := config.NewTemplate("{{.WSID}}")
	spclClr, _ := colors.ParseColor("@special")
	actClr, _ := colors.ParseColor("@active")
	blkClr, _ := colors.ParseColor("@black")
	urgClr, _ := colors.ParseColor("@urgent")

	return Options{
		Format: config.Format{Template: fw},
		Special: SpecialOptions{
			Fg: config.Color(actClr),
			Bg: config.Color(spclClr),
		},
		Active: ActiveOptions{
			Fg: config.Color(blkClr),
			Bg: config.Color(actClr),
		},
		Urgent: UrgentOptions{
			Fg: config.Color(blkClr),
			Bg: config.Color(urgClr),
		},
		OnClick: config.MouseActions[MouseOptions]{
			Actions: map[string]*config.MouseAction[MouseOptions]{},
		},
	}
}
