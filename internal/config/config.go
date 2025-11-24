// Copyright (c) 2025 Nekorg All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// SPDX-License-Identifier: bsd

package config

import (
	"os"

	"github.com/nekorg/pawbar/internal/modules"
	"github.com/nekorg/pawbar/internal/utils"
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
	if err = yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	cfg.Bar.FillDefaults()

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
