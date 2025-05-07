package cpu

import (
	"text/template"
	"time"

	c "github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
)

// Dev NOTE:
// - cpu:
//     tick: 10s
//     cursor: text
//     onclick:
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
	c.RegisterModule("cpu", defaultOptions(), func(o Options) (modules.Module, error) { return &CpuModule{opts: o}, nil })
}

type ThresholdOptions struct {
	Percent c.Percent  `yaml:"percent"`
	For     c.Duration `yaml:"for"`
	Color   c.Color    `yaml:"color"`
}

type Options struct {
	Fg        c.Color                        `yaml:"fg"`
	Bg        c.Color                        `yaml:"bg"`
	Cursor    c.Cursor                       `yaml:"cursor"`
	Tick      c.Duration                     `yaml:"tick"`
	Format    c.Format                       `yaml:"format"`
	Threshold ThresholdOptions               `yaml:"threshold"`
	OnClick   c.OnClickActions[ClickOptions] `yaml:"onclick"`
}

// these field names need to match exactly the
// ones in Options (coz ya know, reflections ;)
// kill me. But they do enable cleaner per-module config code
// so it was worth it.
// Also Rob Pike is such a goated guy
// Read his wise words: https://go.dev/blog/laws-of-reflection
type ClickOptions struct {
	Fg     *c.Color    `yaml:"fg"`
	Bg     *c.Color    `yaml:"bg"`
	Cursor *c.Cursor   `yaml:"cursor"`
	Tick   *c.Duration `yaml:"tick"`
	Format *c.Format   `yaml:"format"`
}

func defaultOptions() Options {
	f, _ := template.New("format").Parse(" {{.Percent}}%")
	urgClr, _ := c.ParseColor("@urgent")
	return Options{
		Format: c.Format{Template: f},
		Tick:   c.Duration(5 * time.Second),
		Threshold: ThresholdOptions{
			Percent: 4,
			For:     c.Duration(3 * time.Second),
			Color:   c.Color(urgClr),
		},
		OnClick: c.OnClickActions[ClickOptions]{},
	}
}
