package config

import (
	"github.com/codelif/pawbar/internal/modules"
	"gopkg.in/yaml.v3"
)

type Factory func(raw *yaml.Node) (modules.Module, error)

var factories = make(map[string]Factory)

func Register(name string, f Factory) { factories[name] = f }
