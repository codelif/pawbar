// Copyright (c) 2025 Nekorg All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// SPDX-License-Identifier: bsd

package config

import (
	"fmt"
	"reflect"

	"dario.cat/mergo"
	"github.com/nekorg/pawbar/internal/modules"
	"gopkg.in/yaml.v3"
)

type Factory func(raw *yaml.Node) (modules.Module, error)

var factories = make(map[string]Factory)

func Register(name string, f Factory) { factories[name] = f }

func RegisterModule[T any](
	name string,
	defaultOpts func() T,
	constructor func(T) (modules.Module, error),
) {
	factories[name] = func(node *yaml.Node) (modules.Module, error) {
		if err := validateOnMouseNode(node); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		opts := defaultOpts()

		if node != nil {
			var userOpts T
			if err := node.Decode(&userOpts); err != nil {
				return nil, fmt.Errorf("%s: bad config: %w", name, err)
			}
			if err := mergo.Merge(&opts, userOpts, mergo.WithOverride); err != nil {
				return nil, fmt.Errorf("%s: merge error: %w", name, err)
			}
		}

		if fv := reflect.ValueOf(&opts).Elem().FieldByName("OnClick"); fv.IsValid() {
			var m reflect.Value

			switch fv.Kind() {
			case reflect.Map:
				m = fv

			case reflect.Struct:
				if sub := fv.FieldByName("Actions"); sub.IsValid() && sub.Kind() == reflect.Map {
					m = sub
				}
			}

			if m.IsValid() {
				for _, key := range m.MapKeys() {
					val := m.MapIndex(key).Interface()
					if v, ok := val.(interface{ Validate() error }); ok {
						if err := v.Validate(); err != nil {
							return nil, fmt.Errorf("%s.onclick.%s: %w",
								name, key.String(), err)
						}
					}
				}
			}
		}

		if v, ok := any(&opts).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
		}

		return constructor(opts)
	}
}
