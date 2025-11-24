// Copyright (c) 2025 Nekorg All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// SPDX-License-Identifier: bsd

package locale

import (
	"time"

	"github.com/nekorg/pawbar/internal/config"
	"github.com/nekorg/pawbar/internal/modules"
)

func init() {
	config.RegisterModule("locale", defaultOptions, func(o Options) (modules.Module, error) { return &LocaleModule{opts: o}, nil })
}

type Options struct {
	Fg      config.Color                      `yaml:"fg"`
	Bg      config.Color                      `yaml:"bg"`
	Cursor  config.Cursor                     `yaml:"cursor"`
	Format  config.Format                     `yaml:"format"`
	Tick    config.Duration                   `yaml:"tick"`
	OnClick config.MouseActions[MouseOptions] `yaml:"onmouse"`
}

type MouseOptions struct {
	Fg     *config.Color    `yaml:"fg"`
	Bg     *config.Color    `yaml:"bg"`
	Cursor *config.Cursor   `yaml:"cursor"`
	Tick   *config.Duration `yaml:"tick"`
	Format *config.Format   `yaml:"format"`
}

func defaultOptions() Options {
	f, _ := config.NewTemplate("{{.Locale}}")
	return Options{
		Format:  config.Format{Template: f},
		Tick:    config.Duration(7 * time.Second),
		OnClick: config.MouseActions[MouseOptions]{},
	}
}
