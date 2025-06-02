package mpris

import (
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
)

func init() {
	config.RegisterModule("mpris", defaultOptions, func(o Options) (modules.Module, error) { return &MprisModule{opts: o}, nil })
}

type PlayOptions struct {
	Icon   rune          `yaml:"icon"`
	Fg     config.Color  `yaml:"fg"`
	Format config.Format `yaml:"format"`
	Bg     config.Color  `yaml:"bg"`
}

type PauseOptions struct {
	Icon   rune          `yaml:"icon"`
	Fg     config.Color  `yaml:"fg"`
	Bg     config.Color  `yaml:"bg"`
	Format config.Format `yaml:"format"`
}

type Options struct {
	Fg      config.Color                      `yaml:"fg"`
	Bg      config.Color                      `yaml:"bg"`
	Cursor  config.Cursor                     `yaml:"cursor"`
	Tick    config.Duration                   `yaml:"tick"`
	Pause   PauseOptions                      `yaml:"pause"`
	Icon    rune                              `yaml:"icon"`
	Play    PlayOptions                       `yaml:"play"`
	Format  config.Format                     `yaml:"format"`
	OnClick config.MouseActions[MouseOptions] `yaml:"onmouse"`
}

type MouseOptions struct {
	Fg     *config.Color    `yaml:"fg"`
	Bg     *config.Color    `yaml:"bg"`
	Cursor *config.Cursor   `yaml:"cursor"`
	Tick   *config.Duration `yaml:"tick"`
	Format *config.Format   `yaml:"format"`
}

func defaultOptions() Options {
	f0, _ := config.NewTemplate("{{.Icon}}")
	f1, _ := config.NewTemplate("{{.Icon}} {{.Artist}} {{.Title}}")
	return Options{
		Icon:   '󰎆',
		Format: config.Format{Template: f0},
		Pause: PauseOptions{
			Icon:   '󰏥',
			Format: config.Format{Template: f1},
		},
		Play: PlayOptions{
			Icon:   '',
			Format: config.Format{Template: f1},
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
