package config

import (
	"gopkg.in/yaml.v3"
)

type Common struct {
	Format string `yaml:"format"`
	Fg     string `yaml:"fg"`
	Bg     string `yaml:"bg"`
}

func Decode(node *yaml.Node, dst any) error {
	if node == nil {
		return nil
	}
	return node.Decode(dst)
}
