package mpris

import (
	"bytes"
	"fmt"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/godbus/dbus/v5"
)

type MprisModule struct {
	receive     chan bool
	send        chan modules.Event
	opts        Options
	initialOpts Options
	artist      string
	title       string
}

func New() modules.MprisModule { return &MprisModule{} }

func (mod *MprisModule) Name() string                                  { return "mpris" }
func (mod *MprisModule) Dependencies() []string                        { return nil }
func (mod *MprisModule) Channels() (<-chan bool, chan<- modules.Event) { return mod.receive, mod.send }

func (mod *MprisModule) Connection() error {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()
}
