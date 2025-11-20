package custom

import (
	"github.com/nekorg/pawbar/internal/config"
	"github.com/nekorg/pawbar/internal/modules"
)

func init() {
	config.RegisterModule("custom", defaultOptions, func(o Options) (modules.Module, error) { return &CustomModule{opts: o}, nil })
}

type Options struct {
	Fg      config.Color                      `yaml:"fg"`
	Bg      config.Color                      `yaml:"bg"`
	Cursor  config.Cursor                     `yaml:"cursor"`
	Tick    config.Duration                   `yaml:"tick"`
	Format  config.Format                     `yaml:"format"`
	OnClick config.MouseActions[MouseOptions] `yaml:"onmouse"`
}

type MouseOptions struct {
	Fg     *config.Color    `yaml:"fg"`
	Bg     *config.Color    `yaml:"bg"`
	Tick   *config.Duration `yaml:"tick"`
	Cursor *config.Cursor   `yaml:"cursor"`
	Format *config.Format   `yaml:"format"`
}

func defaultOptions() Options {
	f, _ := config.NewTemplate("")
	return Options{
		Format:  config.Format{Template: f},
		OnClick: config.MouseActions[MouseOptions]{},
	}
}
