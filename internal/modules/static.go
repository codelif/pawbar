package modules

type StaticModule struct {
	name  string
	cells []EventCell
	deps  []string
}

func NewStaticModule(name string, cells []EventCell, deps []string) *StaticModule {
	return &StaticModule{
		name:  name,
		cells: cells,
		deps:  deps,
	}
}

func (sm *StaticModule) Render() []EventCell {
	return sm.cells
}

func (sm *StaticModule) Run() (<-chan bool, chan<- Event, error) {
	return nil, nil, nil
}

func (sm *StaticModule) Channels() (<-chan bool, chan<- Event) {
	return nil, nil
}

func (sm *StaticModule) Name() string {
	return sm.name
}

func (sm *StaticModule) Dependencies() []string {
	return sm.deps
}
