package config

import (
	"os"

	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/utils"
	"gopkg.in/yaml.v3"
)

func InstantiateModules(cfg *BarConfig) (left, middle, right []modules.Module) {
	left = instantiate(cfg.Left)
	middle = instantiate(cfg.Middle)
	right = instantiate(cfg.Right)

	return left, middle, right
}

func Parse(path string) (*BarConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg BarConfig
	err = yaml.Unmarshal(b, &cfg)
	return &cfg, err
}

func instantiate(specs []ModuleSpec) []modules.Module {
	var out []modules.Module
	for _, s := range specs {
		f, ok := factories[s.Name]
		if !ok {
			utils.Logger.Printf("unknown module '%q'\n", s.Name)
			continue
		}

		m, err := f(s.Params)
		if err != nil {
			utils.Logger.Printf("config error: %v", err)
			continue
		}

		out = append(out, m)
	}

	return out
}
