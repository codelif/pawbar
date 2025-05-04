package config

import (
	"slices"

	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/utils"
)

func LoadModulesFromFile(configPath string) ([]modules.Module, []modules.Module,[]modules.Module, error) {
	barConfig, err := LoadBarConfig(configPath)
	if err != nil {
		return nil, nil,nil, err
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

	var middle []modules.Module
	for _, modName := range barConfig.Middle {
		factory, ok := moduleFactories[modName]
		if !ok {
			utils.Logger.Printf("Unknown module: '%s'\n", modName)
			continue
		}
		middle = append(middle, factory())
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

	return left,middle, right, nil
}

func InitModules(configPath string) (chan modules.Module, []modules.Module,[]modules.Module, []modules.Module, error) {
	tleft,tmiddle, tright, err := LoadModulesFromFile(configPath)
	if err != nil {
		return nil, nil,nil, nil, err
	}

	// runServices(append(tleft, tright...))
	modev := make(chan modules.Module)

	var left []modules.Module
	var middle []modules.Module
	var right []modules.Module

	for _, mod := range tleft {
		rec, _, err := mod.Run()
		if err == nil {
			go runModuleEventLoop(mod, rec, modev)
			left = append(left, mod)
		} else {
			utils.Logger.Println(err)
		}
	}

	for _, mod := range tmiddle {
		rec, _, err := mod.Run()
		if err == nil {
			go runModuleEventLoop(mod, rec, modev)
			middle = append(middle, mod)
		} else {
			utils.Logger.Println(err)
		}
	}

	for _, mod := range tright {
		rec, _, err := mod.Run()
		if err == nil {
			go runModuleEventLoop(mod, rec, modev)
			right = append(right, mod)
		} else {
			utils.Logger.Println(err)
		}
	}

	return modev, left,middle, right, nil
}

func runModuleEventLoop(mod modules.Module, rec <-chan bool, modev chan<- modules.Module) {
	for <-rec {
		modev <- mod
	}
}
