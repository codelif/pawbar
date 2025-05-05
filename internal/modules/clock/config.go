package clock

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
)

// Example config:
//
//  clock:
//     formats:
//       left:  ["%Y-%m-%d %H:%M:%S", "%a %H:%M"]   # left click cycle
//       right: ["%d %B %Y (%A) %H:%M"]             # right click cycle
//     tick:   5s                                   # interval
//
// NOTE: include an example in every module's config.go (also this message)

func init() {
	config.Register("clock", factory)
}

type cfgYaml struct {
	Formats struct {
		Left  []string `yaml:"left"`
		Right []string `yaml:"right"`
	} `yaml:"formats"`
	Tick string `yaml:"tick"`
}

type conf struct {
	left  []string
	right []string
	tick  time.Duration
}

func defaultConfig() conf {
	return conf{
		left: []string{
			"%Y-%m-%d %H:%M:%S",
			"%a %H:%M",
			"%d %B %Y %A %H:%M",
		},
		right: nil,
		tick:  5 * time.Second,
	}
}

func (c *conf) validate() error {
	if len(c.left) == 0 {
		return fmt.Errorf("clock: at least one left format is required")
	}
	return nil
}

func factory(n *yaml.Node) (modules.Module, error) {
	cfg := defaultConfig()

	if n != nil {
		var y cfgYaml
		if err := n.Decode(&y); err != nil {
			return nil, err
		}
		if len(y.Formats.Left) > 0 {
			cfg.left = y.Formats.Left
		}
		if len(y.Formats.Right) > 0 {
			cfg.right = y.Formats.Right
		}
		if y.Tick != "" {
			d, err := time.ParseDuration(y.Tick)
			if err != nil {
				return nil, fmt.Errorf("clock: bad tick duration: %w", err)
			}
			cfg.tick = d
		}
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &ClockModule{cfg: cfg}, nil
}
