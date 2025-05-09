package volume

import (
	"text/template"

	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
)

func init() {
	config.RegisterModule("volume", defaultOptions(), func(o Options) (modules.Module, error) { return &VolumeModule{opts: o}, nil })
}

type Mutedoptions struct {
	MuteFormat string       `yaml:"muteformat"`
	Fg         config.Color `yaml:"fg"`
	Bg         config.Color `yaml:"bg"`
}

type Options struct {
	Fg      config.Color                      `yaml:"fg"`
	Bg      config.Color                      `yaml:"bg"`
	Cursor  config.Cursor                     `yaml:"cursor"`
	Format  config.Format                     `yaml:"format"`
	Muted   Mutedoptions                      `yaml:"muted"`
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
	muteColor, _ := config.ParseColor("darkgray")

	return Options{
		Format: config.Format{Template: fv},
		Icons:  []rune{'󰕿', '󰖀', '󰕾'},
		Muted: Mutedoptions{
			MuteFormat: "󰖁 MUTED",
			Fg:         config.Color(muteColor),
		},
		OnClick: config.MouseActions[MouseOptions]{
			Actions: map[string]*config.MouseAction[MouseOptions]{
				"wheel-up": {
					Run: []string{"pactl", "set-sink-volume", "@DEFAULT_SINK@", "+5%"},
				},
				"wheel-down": {
					Run: []string{"pactl", "set-sink-volume", "@DEFAULT_SINK@", "-5%"},
				},
			},
		},
	}
}
