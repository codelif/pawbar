package wifi

import (
	"time"

	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/lookup/colors"
	"github.com/codelif/pawbar/internal/modules"
)

func init() {
	config.RegisterModule("wifi", defaultOptions, func(o Options) (modules.Module, error) { return &wifiModule{opts: o}, nil })
}

type NoConnectionOptions struct {
	Fg     config.Color  `yaml:"fg"`
	Bg     config.Color  `yaml:"bg"`
	Format config.Format `yaml:"format"`
}

type Options struct {
	Fg           config.Color                      `yaml:"fg"`
	Bg           config.Color                      `yaml:"bg"`
	Cursor       config.Cursor                     `yaml:"cursor"`
	Tick         config.Duration                   `yaml:"tick"`
	Format       config.Format                     `yaml:"format"`
	NoConnection NoConnectionOptions               `yaml:"noconnection"`
	Icons        []rune                            `yaml:"icons"`
	OnClick      config.MouseActions[MouseOptions] `yaml:"onmouse"`
}

type MouseOptions struct {
	Fg     *config.Color    `yaml:"fg"`
	Bg     *config.Color    `yaml:"bg"`
	Cursor *config.Cursor   `yaml:"cursor"`
	Tick   *config.Duration `yaml:"tick"`
	Format *config.Format   `yaml:"format"`
}

func defaultOptions() Options {
	fw, _ := config.NewTemplate("{{.Icon}}")
	fs, _ := config.NewTemplate("{{.Interface}}")
	fn, _ := config.NewTemplate("󰤭")
	fl, _ := config.NewTemplate("{{.Icon}} {{.SSID}}")
	noConClr, _ := colors.ParseColor("darkgray")
	return Options{
		Format: config.Format{Template: fw},
		Tick:   config.Duration(5 * time.Second),
		Icons:  []rune{'󰤯', '󰤟', '󰤢', '󰤥', '󰤨'},
		NoConnection: NoConnectionOptions{
			Format: config.Format{Template: fn},
			Fg:     config.Color(noConClr),
		},
		OnClick: config.MouseActions[MouseOptions]{
			Actions: map[string]*config.MouseAction[MouseOptions]{
				"hover": {
					Configs: []MouseOptions{{Format: &config.Format{Template: fs}}},
				},
				"left": {
					Configs: []MouseOptions{{Format: &config.Format{Template: fl}}},
				},
			},
		},
	}
}
