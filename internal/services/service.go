package services


type Service interface {
    Start() error
    Stop() error
    Name() string
}
var ServiceRegistry = map[string]Service{
    "hypr": &HyprService{}, 
}
