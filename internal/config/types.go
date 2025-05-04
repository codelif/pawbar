package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

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

type BarSettings struct{}

type BarConfig struct {
	Bar    BarSettings  `yaml:"bar"`
	Left   []ModuleSpec `yaml:"left"`
	Middle []ModuleSpec `yaml:"middle"`
	Right  []ModuleSpec `yaml:"right"`
}
