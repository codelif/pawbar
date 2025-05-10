package config

import (
	"fmt"
	"text/template"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"gopkg.in/yaml.v3"
)

type BarSettings struct {
	TruncatePriority []string `yaml:"truncate_priority"`
	Ellipsis         *bool    `yaml:"ellipsis"`
}

func (b *BarSettings) UnmarshalYAML(n *yaml.Node) error {
	type plain BarSettings
	if err := n.Decode((*plain)(b)); err != nil {
		return err
	}

	if len(b.TruncatePriority) != 3 {
		return fmt.Errorf("truncate_priority: exactly 3 anchors needed, %d provided", len(b.TruncatePriority))
	}

	set := map[string]bool{"left": false, "middle": false, "right": false}
	for _, a := range b.TruncatePriority {
		if _, ok := set[a]; !ok {
      return fmt.Errorf(`truncate_priority: invalid anchor %q, valid options are: ["left", "middle", "right"]`, a)
		}
		if set[a] {
			return fmt.Errorf("truncate_priority: %q listed twice", a)
		}
		set[a] = true
	}
	return nil
}

func (b *BarSettings) FillDefaults() {
	if len(b.TruncatePriority) == 0 {
		b.TruncatePriority = []string{"right", "left", "middle"}
	}
	if b.Ellipsis == nil {
		t := true
		b.Ellipsis = &t
	}
}

type BarConfig struct {
	Bar    BarSettings  `yaml:"bar"`
	Left   []ModuleSpec `yaml:"left"`
	Middle []ModuleSpec `yaml:"middle"`
	Right  []ModuleSpec `yaml:"right"`
}

type ModuleSpec struct {
	Name   string
	Params *yaml.Node
}

func (m *ModuleSpec) UnmarshalYAML(n *yaml.Node) error {
	switch n.Kind {
	case yaml.ScalarNode:
		m.Name = n.Value
	case yaml.MappingNode:
		if len(n.Content) != 2 {
			return fmt.Errorf("module mapping must have")
		}
		m.Name = n.Content[0].Value
		m.Params = n.Content[1]
	default:
		return fmt.Errorf("invalid module spec")

	}
	return nil
}

type Duration time.Duration

func (d *Duration) UnmarshalYAML(n *yaml.Node) error {
	var s string
	if err := n.Decode(&s); err != nil {
		return err
	}

	td, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}

	*d = Duration(td)
	return nil
}

func (d Duration) Go() time.Duration { return time.Duration(d) }

type Format struct {
	*template.Template
}

func (t *Format) UnmarshalYAML(n *yaml.Node) error {
	var s string
	if err := n.Decode(&s); err != nil {
		return err
	}

	tmpl, err := template.New("format").Parse(s)
	if err != nil {
		return err
	}

	*t = Format{tmpl}
	return nil
}

func (f Format) Go() *template.Template { return f.Template }

type Color vaxis.Color

func (c *Color) UnmarshalYAML(n *yaml.Node) error {
	var s string
	if err := n.Decode(&s); err != nil {
		return err
	}

	col, err := ParseColor(s)
	if err != nil {
		return err
	}
	*c = Color(col)

	return nil
}

func (c Color) Go() vaxis.Color { return vaxis.Color(c) }

type Percent int

func (p *Percent) UnmarshalYAML(n *yaml.Node) error {
	var s int
	if err := n.Decode(&s); err != nil {
		return err
	}

	if s < 0 || s > 100 {
		return fmt.Errorf("percentage should be between 0-100")
	}
	*p = Percent(s)
	return nil
}

func (p Percent) Go() int { return int(p) }

type Cursor vaxis.MouseShape

func (c *Cursor) UnmarshalYAML(n *yaml.Node) error {
	var s string
	if err := n.Decode(&s); err != nil {
		return err
	}

	cur, err := ParseCursor(s)
	if err != nil {
		return err
	}

	*c = Cursor(cur)
	return nil
}

func (c Cursor) Go() vaxis.MouseShape { return vaxis.MouseShape(c) }
