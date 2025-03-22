package main


type Service interface {
    Start() error
    Stop() error
    Name() string
}
var serviceRegistry = map[string]Service{
    "hypr": &HyprService{}, 
}
