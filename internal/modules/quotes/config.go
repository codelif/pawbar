package quotes

import (
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
)

func init() {
	config.RegisterModule("quotes", defaultOptions, func(o Options) (modules.Module, error) { return &Quotes{opts: o}, nil })
}

type Options struct {
	Fg      config.Color                      `yaml:"fg"`
	Bg      config.Color                      `yaml:"bg"`
	Cursor  config.Cursor                     `yaml:"cursor"`
	Format  config.Format                     `yaml:"format"`
	OnClick config.MouseActions[MouseOptions] `yaml:"onmouse"`
}

type MouseOptions struct {
	Fg     *config.Color  `yaml:"fg"`
	Bg     *config.Color  `yaml:"bg"`
	Cursor *config.Cursor `yaml:"cursor"`
	Format *config.Format `yaml:"format"`
}

func defaultOptions() Options {
	f, _ := config.NewTemplate("{{.Quote}}")
	return Options{
		Format:  config.Format{Template: f},
		OnClick: config.MouseActions[MouseOptions]{},
	}
}
