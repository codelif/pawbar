package config

import (
	"fmt"
	"reflect"

	"dario.cat/mergo"
	"github.com/codelif/pawbar/internal/modules"
	"gopkg.in/yaml.v3"
)

type Factory func(raw *yaml.Node) (modules.Module, error)

var factories = make(map[string]Factory)

func Register(name string, f Factory) { factories[name] = f }

func RegisterModule[T any](
	name string,
	defaultOpts T,
	constructor func(T) (modules.Module, error),
) {
	factories[name] = func(node *yaml.Node) (modules.Module, error) {
		if err := validateOnClickNode(node); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		opts := defaultOpts

		if node != nil {
			var userOpts T
			if err := node.Decode(&userOpts); err != nil {
				return nil, fmt.Errorf("%s: bad config: %w", name, err)
			}
			if err := mergo.Merge(&opts, userOpts, mergo.WithOverride); err != nil {
				return nil, fmt.Errorf("%s: merge error: %w", name, err)
			}
		}

		if f := reflect.ValueOf(&opts).Elem().FieldByName("OnClick"); f.IsValid() {
			for _, key := range f.MapKeys() {
				act, ok := f.MapIndex(key).Interface().(interface{ Validate() error })
				if ok {
					if err := act.Validate(); err != nil {
						return nil, fmt.Errorf("%s.onclick.%s: %w", name, key.String(), err)
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
