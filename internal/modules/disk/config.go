package disk

import (
	"time"

	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/lookup/colors"
	"github.com/codelif/pawbar/internal/lookup/icons"
	"github.com/codelif/pawbar/internal/modules"
)

func init() {
	config.RegisterModule("disk", defaultOptions, func(o Options) (modules.Module, error) { return &DiskModule{opts: o}, nil })
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

type Options struct {
	Fg     config.Color    `yaml:"fg"`
	Bg     config.Color    `yaml:"bg"`
	Cursor config.Cursor   `yaml:"cursor"`
	Tick   config.Duration `yaml:"tick"`
	Format config.Format   `yaml:"format"`
	Icon   config.Icon     `yaml:"icon"`

	UseSI bool         `yaml:"use_si"`
	Scale config.Scale `yaml:"unit"`

	Warning WarningOptions `yaml:"warning"`
	Urgent  UrgentOptions  `yaml:"urgent"`

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
		Warning: WarningOptions{
			Percent: 90,
			Fg:      config.Color(warClr),
		},
		Urgent: UrgentOptions{
			Percent: 95,
			Fg:      config.Color(urgClr),
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
