package cpu

import (
	"text/template"
	"time"

	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
)

// Dev NOTE:
// - cpu:
//     tick: 10s
//     cursor: text
//     onmouse:
//       left:
//         notify: "wow"
//         run: ["pavucontrol"]
//         config:
//           - cursor: pointer
//             format: wow
//           - format: "{{.Percent}} this is shit"
//             fg: yellow
//             bg: blue
//           - tick: 1s
//
// In config:, it can be a list like above or just a single alternative:
// left:
//   config:
//     fg: aliceblue
//     format: hello
//
// this will just alternate between this state and initial state
// also note fields not defined (like bg, cursor, tick in above example)
// use their initial values they don't carry from previous alternate
//
// all of this will be true for all options with config.OnClickActions field
// (they also need to call the relevent OnClickAction function in the loop)
// see internal/modules/cpu/cpu.go and internal/config/click.go for more info

// you can also attach a Validate function to Options but try to avoid if you
// can define a yaml type in internal/config/types.go

func init() {
	config.RegisterModule("cpu", defaultOptions, func(o Options) (modules.Module, error) { return &CpuModule{opts: o}, nil })
}

type ThresholdOptions struct {
	Percent config.Percent  `yaml:"percent"`
	For     config.Duration `yaml:"for"`
	Fg      config.Color    `yaml:"fg"`
	Bg      config.Color    `yaml:"bg"`
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

// these field names need to match exactly the
// ones in Options (coz ya know, reflections ;)
// kill me. But they do enable cleaner per-module config code
// so it was worth it.
// Also Rob Pike is such a goated guy
// Read his wise words: https://go.dev/blog/laws-of-reflection
type MouseOptions struct {
	Fg     *config.Color    `yaml:"fg"`
	Bg     *config.Color    `yaml:"bg"`
	Cursor *config.Cursor   `yaml:"cursor"`
	Tick   *config.Duration `yaml:"tick"`
	Format *config.Format   `yaml:"format"`
}

func defaultOptions() Options {
	f, _ := template.New("format").Parse(" {{.Percent}}%")
	urgClr, _ := config.ParseColor("@urgent")
	return Options{
		Format: config.Format{Template: f},
		Tick:   config.Duration(3 * time.Second),
		Threshold: ThresholdOptions{
			Percent: 90,
			For:     config.Duration(7 * time.Second),
			Fg:      config.Color(urgClr),
		},
		OnClick: config.MouseActions[MouseOptions]{},
	}
}
