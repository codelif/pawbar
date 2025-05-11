package clock

import (
	"time"

	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
)

// Example config:
//
//  clock:
//     format: "%Y-%m-%d %H:%M:%S"
//     tick:   5s                                   # interval
//     onmouse:
//       left:
//         config:
//           format: "%a %H:%M"
//       right:
//         config:
//           format: "%d %B %Y (%A) %H:%M"
//
// NOTE: include an example in every module's config.go (also this message)

func init() {
	config.RegisterModule("clock", defaultOptions, func(o Options) (modules.Module, error) { return &ClockModule{opts: o}, nil })
}

type Options struct {
	Fg      config.Color                      `yaml:"fg"`
	Bg      config.Color                      `yaml:"bg"`
	Cursor  config.Cursor                     `yaml:"cursor"`
	Tick    config.Duration                   `yaml:"tick"`
	Format  string                            `yaml:"format"`
	OnClick config.MouseActions[MouseOptions] `yaml:"onmouse"`
}

type MouseOptions struct {
	Fg     *config.Color    `yaml:"fg"`
	Bg     *config.Color    `yaml:"bg"`
	Cursor *config.Cursor   `yaml:"cursor"`
	Tick   *config.Duration `yaml:"tick"`
	Format *string          `yaml:"format"`
}

func defaultOptions() Options {
	return Options{
		Format:  "%Y-%m-%d %H:%M:%S",
		Tick:    config.Duration(5 * time.Second),
		OnClick: config.MouseActions[MouseOptions]{},
	}
}
