// Copyright (c) 2025 Nekorg All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// SPDX-License-Identifier: bsd

package bluetooth

import (
	"github.com/nekorg/pawbar/internal/config"
	"github.com/nekorg/pawbar/internal/lookup/colors"
	"github.com/nekorg/pawbar/internal/modules"
)

func init() {
	config.RegisterModule("bluetooth", defaultOptions, func(o Options) (modules.Module, error) { return &bluetoothModule{opts: o}, nil })
}

type NoConnectionOptions struct {
	Fg     config.Color  `yaml:"fg"`
	Bg     config.Color  `yaml:"bg"`
	Format config.Format `yaml:"format"`
}

type ConnectionOptions struct {
	Fg     config.Color  `yaml:"fg"`
	Bg     config.Color  `yaml:"bg"`
	Format config.Format `yaml:"format"`
}

type Options struct {
	Fg           config.Color                      `yaml:"fg"`
	Bg           config.Color                      `yaml:"bg"`
	Cursor       config.Cursor                     `yaml:"cursor"`
	Format       config.Format                     `yaml:"format"`
	Connection   ConnectionOptions                 `yaml:"connection"`
	NoConnection NoConnectionOptions               `yaml:"noconnection"`
	OnClick      config.MouseActions[MouseOptions] `yaml:"onmouse"`
}

type MouseOptions struct {
	Fg     *config.Color  `yaml:"fg"`
	Bg     *config.Color  `yaml:"bg"`
	Cursor *config.Cursor `yaml:"cursor"`
	Format *config.Format `yaml:"format"`
}

func defaultOptions() Options {
	fd, _ := config.NewTemplate("󰂱")
	fc, _ := config.NewTemplate("")
	fn, _ := config.NewTemplate("󰂲")
	fa, _ := config.NewTemplate("󰂱 {{.Device}}")
	noConClr, _ := colors.ParseColor("darkgray")
	return Options{
		Format: config.Format{Template: fd},
		NoConnection: NoConnectionOptions{
			Format: config.Format{Template: fn},
			Fg:     config.Color(noConClr),
		},
		Connection: ConnectionOptions{
			Format: config.Format{Template: fc},
		},
		OnClick: config.MouseActions[MouseOptions]{
			Actions: map[string]*config.MouseAction[MouseOptions]{
				"left": {
					Configs: []MouseOptions{{Format: &config.Format{Template: fa}}},
				},
			},
		},
	}
}
