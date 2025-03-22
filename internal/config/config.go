package config

import (
	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/services"
	"github.com/codelif/pawbar/internal/utils"
)

func LoadModules() ([]modules.Module, []modules.Module) {
	left_gen := [](func() modules.Module){modules.NewHyprWorkspaces, modules.NewHyprTitle}
	right_gen := []func() modules.Module{modules.NewClock}

	var left, right []modules.Module

	for _, genmod := range left_gen {
		left = append(left, genmod())
	}
	for _, genmod := range right_gen {
		right = append(right, genmod())
	}

	return left, right
}

func InitModules() (chan modules.Module, []modules.Module, []modules.Module) {
	tleft, tright := LoadModules()
	RunServices(append(tleft, tright...))
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

	return modev, left, right
}

func runModuleEventLoop(mod modules.Module, rec <-chan bool, modev chan<- modules.Module) {
	for <-rec {
		modev <- mod
	}
}

func RunServices(mods []modules.Module) {
	neededServices := make(map[string]bool)
	for _, mod := range mods {
		for _, dep := range mod.Dependencies() {
			neededServices[dep] = true
		}
	}

	// 2) Start each needed service
	for dep := range neededServices {
		svc, ok := services.ServiceRegistry[dep]
		if !ok {
			utils.Logger.Printf("Unknown service dependency: %s\n", dep)
			continue
		}
		svc.Start() // handle errors etc.
	}
}

