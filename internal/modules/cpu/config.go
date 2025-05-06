package cpu

import (
	"fmt"
	"text/template"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	c "github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
	mouse "github.com/codelif/pawbar/internal/mouse"
	"gopkg.in/yaml.v3"
)

func init() {
	c.Register("cpu", factory)
}

type ThresholdOptions struct {
	Percent int    `yaml:"percent"`
	For     string `yaml:"for"`
	Color   string `yaml:"color"`
}

type Options struct {
	c.Common  `yaml:",inline"`
	Cursor    string             `yaml:"cursor"`
	Tick      string             `yaml:"tick"`
	Threshold ThresholdOptions   `yaml:"threshold"`
	OnClicks  mouse.Map[Options] `yaml:"on-click"`
}

func (o Options) GetOnClicks() mouse.Map[Options] { return o.OnClicks }
func (o Options) GetCursor() string { return o.Cursor }

func defaultOptions() Options {
	return Options{
		Common: c.Common{
			Format: " {{.Percent}}%",
			Fg:     "@default",
			Bg:     "@default",
		},
		Cursor: "default",
		Tick:   "5s",
		Threshold: ThresholdOptions{
			Percent: 90,
			For:     "3s",
			Color:   "@urgent",
		},
	}
}

func (o *Options) Validate() error {
	if o.Format != "" {
		if _, err := template.New("format").Parse(o.Format); err != nil {
			return err
		}
	}

	if o.Tick != "" {
		if _, err := time.ParseDuration(o.Tick); err != nil {
			return fmt.Errorf("cpu: bad tick: %w", err)
		}
	}

	if o.Threshold.Percent < 0 || o.Threshold.Percent > 100 {
		return fmt.Errorf("cpu: threshold.percent must be 0–100")
	}

	if o.Threshold.For != "" {
		if _, err := time.ParseDuration(o.Threshold.For); err != nil {
			return fmt.Errorf("cpu: threshold.for: %w", err)
		}
	}

	if _, err := c.ParseColor(o.Threshold.Color); err != nil {
		return err
	}

	if _, err := c.ParseColor(o.Fg); err != nil {
		return err
	}

	if _, err := c.ParseColor(o.Bg); err != nil {
		return err
	}

	for btn, h := range o.OnClicks {
		if h.Config != nil {
			if err := h.Config.Validate(); err != nil {
				return fmt.Errorf("cpu: on-click.%s: %w", btn, err)
			}
		}
		if h.IsEmpty() {
			return fmt.Errorf("cpu: on-click.%s is empty", btn)
		}
	}

	return nil
}

type Config struct {
	Tmpl             *template.Template
	FgColor, BgColor vaxis.Color
	Tick             time.Duration

	ThrPercent int
	ThrFor     time.Duration
	ThrColor   vaxis.Color

	Cursor   vaxis.MouseShape
	OnClicks mouse.Map[Options]
}

// does not validate options
func newConfig(o Options) Config {
	tmpl, _ := template.New("cpu").Parse(o.Format)
	tick, _ := time.ParseDuration(o.Tick)
	forDur, _ := time.ParseDuration(o.Threshold.For)

	fgCol, _ := c.ParseColor(o.Fg)
	bgCol, _ := c.ParseColor(o.Bg)
	thrCol, _ := c.ParseColor(o.Threshold.Color)

	return Config{
		Tmpl:       tmpl,
		Tick:       tick,
		FgColor:    fgCol,
		BgColor:    bgCol,
		ThrColor:   thrCol,
		ThrPercent: o.Threshold.Percent,
		ThrFor:     forDur,
    Cursor: mouse.ParseCursor(o.Cursor),
		OnClicks:   o.OnClicks,
	}
}

func factory(n *yaml.Node) (modules.Module, error) {
	opts := defaultOptions()
	if err := c.Decode(n, &opts); err != nil {
		return nil, err
	}
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	cfg := newConfig(opts)

	return &CpuModule{
		opts: opts,
		cfg:  cfg,
	}, nil
}
