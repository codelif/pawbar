// Copyright (c) 2025 Nekorg All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// SPDX-License-Identifier: bsd

package config

import (
	"fmt"
	"text/template"
)

func Funcs() template.FuncMap {
	return template.FuncMap{
		"round": func(p int, v interface{}) string {
			switch x := v.(type) {
			case float32, float64:
				return fmt.Sprintf("%.*f", p, x)
			default:
				return fmt.Sprintf("%v", v)
			}
		},
	}
}

func NewTemplate(src string) (*template.Template, error) {
	return template.New("format").Funcs(Funcs()).Parse(src)
}
