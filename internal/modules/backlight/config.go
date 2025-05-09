package backlight

import (
	"text/template"

	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
)

func init() {
	config.RegisterModule("backlight", defaultOptions(), func(o Options) (modules.Module, error) { return &Backlight{opts: o}, nil })
}

type Options struct {
	Fg      config.Color                      `yaml:"fg"`
	Bg      config.Color                      `yaml:"bg"`
	Cursor  config.Cursor                     `yaml:"cursor"`
	Format  config.Format                     `yaml:"format"`
	Icons   []rune                            `yaml:"icons"`
	OnClick config.MouseActions[MouseOptions] `yaml:"onmouse"`
}

type MouseOptions struct {
	Fg     *config.Color  `yaml:"fg"`
	Bg     *config.Color  `yaml:"bg"`
	Cursor *config.Cursor `yaml:"cursor"`
	Format *config.Format `yaml:"format"`
}

func defaultOptions() Options {
	fv, _ := template.New("format").Parse("{{.Icon}} {{.Percent}}%")

	return Options{
		Format: config.Format{Template: fv},
		Icons:  []rune{'󰃞', '󰃟', '󰃝', '󰃠'},
		OnClick: config.MouseActions[MouseOptions]{
			Actions: map[string]*config.MouseAction[MouseOptions]{
				"wheel-up": {
					Run: []string{"brightnessctl", "set", "+5%"},
				},
				"wheel-down": {
					Run: []string{"brightnessctl", "set", "-5%"},
				},
			},
		},
	}
}
