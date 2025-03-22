package services

type Service interface {
	Start() error
	Stop() error
	Name() string
}

var ServiceRegistry = make(map[string]Service)

func StartService(name string, s Service) {
	prevService, ok := ServiceRegistry[name]
	if ok {
		prevService.Stop()
	}

	ServiceRegistry[name] = s
	s.Start()
}
