package disk

import (
	"time"

	"github.com/nekorg/pawbar/internal/config"
	"github.com/nekorg/pawbar/internal/lookup/colors"
	"github.com/nekorg/pawbar/internal/lookup/icons"
	"github.com/nekorg/pawbar/internal/modules"
)

func init() {
	config.RegisterModule("disk", defaultOptions, func(o Options) (modules.Module, error) { return &DiskModule{opts: o}, nil })
}

type ThresholdOptions struct {
	Percent   config.Percent   `yaml:"percent"`
	Direction config.Direction `yaml:"direction"`
	Fg        config.Color     `yaml:"fg"`
	Bg        config.Color     `yaml:"bg"`
}

type Options struct {
	Fg     config.Color    `yaml:"fg"`
	Bg     config.Color    `yaml:"bg"`
	Cursor config.Cursor   `yaml:"cursor"`
	Tick   config.Duration `yaml:"tick"`
	Format config.Format   `yaml:"format"`
	Icon   config.Icon     `yaml:"icon"`

	UseSI bool         `yaml:"use_si"`
	Scale config.Scale `yaml:"unit"`

	Thresholds []ThresholdOptions `yaml:"thresholds"`

	OnClick config.MouseActions[MouseOptions] `yaml:"onmouse"`
}

type MouseOptions struct {
	Fg     *config.Color    `yaml:"fg"`
	Bg     *config.Color    `yaml:"bg"`
	Cursor *config.Cursor   `yaml:"cursor"`
	Tick   *config.Duration `yaml:"tick"`
	Format *config.Format   `yaml:"format"`
	Icon   *config.Icon     `yaml:"icon"`

	UseSI *bool         `yaml:"use_si"`
	Scale *config.Scale `yaml:"scale"`
}

func defaultOptions() Options {
	icon, _ := icons.Lookup("disk")
	f0, _ := config.NewTemplate("{{.Icon}} {{.UsedPercent}}%")
	f1, _ := config.NewTemplate("{{.Icon}} {{.Used | round 2}}/{{.Total | round 2}} {{.Unit}}")
	urgClr, _ := colors.ParseColor("@urgent")
	warClr, _ := colors.ParseColor("@warning")
	return Options{
		Format: config.Format{Template: f0},
		Tick:   config.Duration(10 * time.Second),
		UseSI:  false,
		Icon:   config.Icon(icon),
		Thresholds: []ThresholdOptions{
			{
				Percent:   80,
				Direction: config.Direction(true),
				Fg:        config.Color(warClr),
			},
			{
				Percent:   90,
				Direction: config.Direction(true),
				Fg:        config.Color(urgClr),
			},
		},
		OnClick: config.MouseActions[MouseOptions]{
			Actions: map[string]*config.MouseAction[MouseOptions]{
				"left": {
					Configs: []MouseOptions{{Format: &config.Format{Template: f1}}},
				},
			},
		},
	}
}
