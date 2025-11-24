// Copyright (c) 2025 Nekorg All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// SPDX-License-Identifier: bsd

package icons

import (
	"fmt"
	"regexp"

	"github.com/nekorg/pawbar/internal/utils"
)

var table = map[string]string{
	"disk":    "",
	"compass": "",
}

func Register(name, glyph string) {
	table[name] = glyph
}

func Lookup(name string) (string, error) {
	g, ok := table[name]
	if !ok {
		return "", fmt.Errorf("unknown icon: %q", name)
	}
	return g, nil
}

var re = regexp.MustCompile(`@[@A-Za-z0-9_]+`)

func Resolve(s string) string {
	return re.ReplaceAllStringFunc(s, func(m string) string {
		if m == "@@" {
			return "@"
		}
		if g, ok := table[m[1:]]; ok {
			return g
		}
		return m
	})
}

// linearly chooses an icon from a sorted list based on percent
func Choose(icons []rune, percent int) rune {
	return icons[utils.Clamp((len(icons)-1)*percent/100, 0, len(icons)-1)]
}
