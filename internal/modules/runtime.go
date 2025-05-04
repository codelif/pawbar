package modules

import "github.com/codelif/pawbar/internal/utils"

func Init(left, middle, right []Module) (
	chan Module, // modev
	[]Module, // left
	[]Module, // middle
	[]Module, // right
) {
	modev := make(chan Module)
	return modev, startModules(left, modev), startModules(middle, modev), startModules(right, modev)
}

func startModules(mods []Module, modev chan<- Module) []Module {
	var okMods []Module

	for _, m := range mods {
		rec, _, err := m.Run()
		if err != nil {
			utils.Logger.Printf("error starting module '%s': %v\n", m.Name(), err)
			continue
		}

		go func(m Module, rec <-chan bool) {
			for range rec {
				modev <- m
			}
		}(m, rec)
		okMods = append(okMods, m)
	}
	return okMods
}
