package config

import (
	"slices"

	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/modules/clock"
	"github.com/codelif/pawbar/internal/modules/hyprtitle"
	"github.com/codelif/pawbar/internal/modules/hyprws"
	"github.com/codelif/pawbar/internal/services/hypr"
	"github.com/codelif/pawbar/internal/utils"
)

func LoadModulesFromFile(configPath string) ([]modules.Module, []modules.Module, error) {
	barConfig, err := LoadBarConfig(configPath)
	if err != nil {
		return nil, nil, err
	}

	var left []modules.Module
	for _, modName := range barConfig.Left {
		factory, ok := moduleFactories[modName]
		if !ok {
			utils.Logger.Printf("Unknown module: '%s'\n", modName)
			continue
		}
		left = append(left, factory())
	}

	// Build right modules
	var right []modules.Module
	for _, modName := range barConfig.Right {
		factory, ok := moduleFactories[modName]
		if !ok {
			utils.Logger.Printf("Unknown module: '%s'\n", modName)
			continue
		}
		right = append(right, factory())
	}
	slices.Reverse(right)
	return left, right, nil
}

func LoadModules() ([]modules.Module, []modules.Module) {
	left_gen := [](func() modules.Module){hyprws.New, hyprtitle.New}
	right_gen := []func() modules.Module{clock.New}

	var left, right []modules.Module

	for _, genmod := range left_gen {
		left = append(left, genmod())
	}
	for _, genmod := range right_gen {
		right = append(right, genmod())
	}

	return left, right
}

func InitModules(configPath string) (chan modules.Module, []modules.Module, []modules.Module, error) {
	tleft, tright, err := LoadModulesFromFile(configPath)
	if err != nil {
		return nil, nil, nil, err
	}

	runServices(append(tleft, tright...))
	modev := make(chan modules.Module)

	var left []modules.Module
	var right []modules.Module

	for _, mod := range tleft {
		rec, _, err := mod.Run()
		if err == nil {
			go runModuleEventLoop(mod, rec, modev)
			left = append(left, mod)
		}
	}

	for _, mod := range tright {
		rec, _, err := mod.Run()
		if err == nil {
			go runModuleEventLoop(mod, rec, modev)
			right = append(right, mod)
		}
	}

	return modev, left, right, nil
}

func runModuleEventLoop(mod modules.Module, rec <-chan bool, modev chan<- modules.Module) {
	for <-rec {
		modev <- mod
	}
}

func runServices(mods []modules.Module) {
	neededServices := make(map[string]bool)
	for _, mod := range mods {
		for _, dep := range mod.Dependencies() {
			neededServices[dep] = true
		}
	}

	for dep := range neededServices {
		switch dep {
		case "hypr":
			hypr.Register()
		default:
		}
	}
}
