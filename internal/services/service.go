package services

import "github.com/codelif/pawbar/internal/utils"

type Service interface {
	Start() error
	Stop() error
	Name() string
}

var ServiceRegistry = make(map[string]Service)

func Ensure(name string, factory func() Service) Service {
	if s, ok := ServiceRegistry[name]; ok {
		return s
	}

	s := factory()
	StartService(name, s)
	return s
}

func StartService(name string, s Service) {
	prevService, ok := ServiceRegistry[name]
	if ok {
    utils.Logger.Printf("services: stopping '%s'\n", s.Name())
		prevService.Stop()
	}
  
  utils.Logger.Printf("services: starting '%s'\n", s.Name())
	ServiceRegistry[name] = s
	s.Start()
}
