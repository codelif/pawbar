package battery

import (
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/lookup/colors"
	"github.com/codelif/pawbar/internal/modules"
)

func init() {
	config.RegisterModule("battery", defaultOptions, func(o Options) (modules.Module, error) { return &Battery{opts: o}, nil })
}

type ThresholdOptions struct {
	Percent   config.Percent   `yaml:"percent"`
	Direction config.Direction `yaml:"direction"`
	Fg        config.Color     `yaml:"fg"`
	Bg        config.Color     `yaml:"bg"`
}

type DischargingOptions struct {
	Fg    config.Color `yaml:"fg"`
	Bg    config.Color `yaml:"bg"`
	Icons []rune       `yaml:"icons"`
}

type ChargingOptions struct {
	Fg    config.Color `yaml:"fg"`
	Bg    config.Color `yaml:"bg"`
	Icons []rune       `yaml:"icons"`
}

type ChargedOptions struct {
	Fg   config.Color `yaml:"fg"`
	Bg   config.Color `yaml:"bg"`
	Icon rune         `yaml:"icon"`
}

type Options struct {
	Fg          config.Color                      `yaml:"fg"`
	Bg          config.Color                      `yaml:"bg"`
	Cursor      config.Cursor                     `yaml:"cursor"`
	Format      config.Format                     `yaml:"format"`
	Discharging DischargingOptions                `yaml:"discharging"`
	Charging    ChargingOptions                   `yaml:"charging"`
	Charged     ChargedOptions                    `yaml:"charged"`
	Thresholds  []ThresholdOptions                `yaml:"thresholds"`
	OnClick     config.MouseActions[MouseOptions] `yaml:"onmouse"`
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
	return Options{
		Format: config.Format{Template: fv},
		Discharging: DischargingOptions{
			Icons: []rune{'󰂃', '󰁺', '󰁻', '󰁼', '󰁽', '󰁾', '󰁿', '󰂀', '󰂁', '󰂂', '󰁹'},
		},
		Charging: ChargingOptions{
			Icons: []rune{'󰢟', '󰢜', '󰂆', '󰂇', '󰂈', '󰢝', '󰂉', '󰢞', '󰂊', '󰂋', '󰂅'},
		},
		Charged: ChargedOptions{
			Icon: '󱟢',
		},
		Thresholds: []ThresholdOptions{
			{
				Percent: 15,
				Fg:      config.Color(urgClr),
			},
			{
				Percent: 30,
				Fg:      config.Color(warClr),
			},
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
