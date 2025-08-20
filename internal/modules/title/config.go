package title

import (
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/lookup/colors"
	"github.com/codelif/pawbar/internal/modules"
)

func init() {
	config.RegisterModule("title", defaultOptions, func(o Options) (modules.Module, error) { return &Module{opts: o}, nil })
}

type DataOptions struct {
	Format config.Format `yaml:"format"`
	Fg     config.Color  `yaml:"fg"`
	Bg     config.Color  `yaml:"bg"`
}

type Options struct {
	Fg      config.Color                      `yaml:"fg"`
	Bg      config.Color                      `yaml:"bg"`
	Cursor  config.Cursor                     `yaml:"cursor"`
	Title   DataOptions                       `yaml:"title"`
	Class   DataOptions                       `yaml:"class"`
	OnClick config.MouseActions[MouseOptions] `yaml:"onmouse"`
}

type MouseOptions struct {
	Fg     *config.Color  `yaml:"fg"`
	Bg     *config.Color  `yaml:"bg"`
	Cursor *config.Cursor `yaml:"cursor"`
}

func defaultOptions() Options {
	fc, _ := config.NewTemplate("{{.Class}}")
	ft, _ := config.NewTemplate("{{.Title}}")
	clClr, _ := colors.ParseColor("@cool")
	blkClr, _ := colors.ParseColor("@black")

	return Options{
		Title: DataOptions{
			Format: config.Format{Template: ft},
		},
		Class: DataOptions{
			Format: config.Format{Template: fc},
			Bg:     config.Color(clClr),
			Fg:     config.Color(blkClr),
		},
		OnClick: config.MouseActions[MouseOptions]{
			Actions: map[string]*config.MouseAction[MouseOptions]{},
		},
	}
}
