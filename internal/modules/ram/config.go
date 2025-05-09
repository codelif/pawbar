package ram

import (
	"text/template"
	"time"

	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
)

func init() {
	config.RegisterModule("ram", defaultOptions(), func(o Options) (modules.Module, error) { return &RamModule{opts: o}, nil })
}

type ThresholdOptions struct {
	PercentUrg config.Percent `yaml:"percentUrgent"`
	PercentWar config.Percent `yaml:"percentWarning"`
	FgWar      config.Color   `yaml:"fgWarning"`
	BgWar      config.Color   `yaml:"bgWarning"`
	FgUrg      config.Color   `yaml:"fgUrgent"`
	BgUrg      config.Color   `yaml:"bgUrgent"`
}

type Options struct {
	Fg        config.Color                      `yaml:"fg"`
	Bg        config.Color                      `yaml:"bg"`
	Cursor    config.Cursor                     `yaml:"cursor"`
	Tick      config.Duration                   `yaml:"tick"`
	Format    config.Format                     `yaml:"format"`
	Threshold ThresholdOptions                  `yaml:"threshold"`
	OnClick   config.MouseActions[MouseOptions] `yaml:"onmouse"`
}

type MouseOptions struct {
	Fg     *config.Color    `yaml:"fg"`
	Bg     *config.Color    `yaml:"bg"`
	Cursor *config.Cursor   `yaml:"cursor"`
	Tick   *config.Duration `yaml:"tick"`
	Format *config.Format   `yaml:"format"`
}

func defaultOptions() Options {
	f, _ := template.New("format").Parse("󰆌 {{.Percent}}%")
	urgClr, _ := config.ParseColor("@urgent")
	warClr, _ := config.ParseColor("@warning")
	return Options{
		Format: config.Format{Template: f},
		Threshold: ThresholdOptions{
			PercentUrg: 90,
			PercentWar: 80,
			FgUrg:      config.Color(urgClr),
			FgWar:      config.Color(warClr),
		},
		OnClick: config.MouseActions[MouseOptions]{},
	}
}
