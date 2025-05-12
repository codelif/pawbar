package battery

import (
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/lookup/colors"
	"github.com/codelif/pawbar/internal/modules"
)

func init() {
	config.RegisterModule("battery", defaultOptions, func(o Options) (modules.Module, error) { return &Battery{opts: o}, nil })
}

type WarningOptions struct {
	Percent config.Percent `yaml:"percent"`
	Fg      config.Color   `yaml:"fg"`
	Bg      config.Color   `yaml:"bg"`
}

type UrgentOptions struct {
	Percent config.Percent `yaml:"percent"`
	Fg      config.Color   `yaml:"fg"`
	Bg      config.Color   `yaml:"bg"`
}

type OptimalOptions struct {
	Percent config.Percent `yaml:"percent"`
	Fg      config.Color   `yaml:"fg"`
	Bg      config.Color   `yaml:"bg"`
}

type Options struct {
	Fg               config.Color                      `yaml:"fg"`
	Bg               config.Color                      `yaml:"bg"`
	Cursor           config.Cursor                     `yaml:"cursor"`
	Format           config.Format                     `yaml:"format"`
	FormatTimeRem    config.Format                     `yaml:"formatTimeRem"`
	IconsDischarging []rune                            `yaml:"iconsDischarging"`
	IconsCharging    []rune                            `yaml:"iconsCharging"`
	Warning          WarningOptions                    `yaml:"warning"`
	Urgent           UrgentOptions                     `yaml:"urgent"`
	Optimal          OptimalOptions                    `yaml:"optimal"`
	OnClick          config.MouseActions[MouseOptions] `yaml:"onmouse"`
}

type MouseOptions struct {
	Fg     *config.Color  `yaml:"fg"`
	Bg     *config.Color  `yaml:"bg"`
	Cursor *config.Cursor `yaml:"cursor"`
	Format *config.Format `yaml:"format"`
}

func defaultOptions() Options {
	fv, _ := config.NewTemplate("{{.Icon}} {{.Percent}}%")
	fr, _ := config.NewTemplate("{{.Hours}} hrs {{ .Minutes}} mins")
	urgClr, _ := colors.ParseColor("@urgent")
	warClr, _ := colors.ParseColor("@warning")
	optClr, _ := colors.ParseColor("@good")
	return Options{
		Format:           config.Format{Template: fv},
		IconsDischarging: []rune{'󰂃', '󰁺', '󰁻', '󰁼', '󰁽', '󰁾', '󰁿', '󰂀', '󰂁', '󰂂', '󰁹'},
		IconsCharging:    []rune{'󰢟', '󰢜', '󰂆', '󰂇', '󰂈', '󰢝', '󰂉', '󰢞', '󰂊', '󰂋', '󰂅'},
		Urgent: UrgentOptions{
			Percent: 15,
			Fg:      config.Color(urgClr),
		},
		Warning: WarningOptions{
			Percent: 30,
			Fg:      config.Color(warClr),
		},
		Optimal: OptimalOptions{
			Percent: 90,
			Fg:      config.Color(optClr),
		},
		OnClick: config.MouseActions[MouseOptions]{
			Actions: map[string]*config.MouseAction[MouseOptions]{
				"hover": {
					Configs: []MouseOptions{{Format: &config.Format{Template: fr}}},
				},
			},
		},
	}
}
