package bluetooth

import (
	"bytes"
	"fmt"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/godbus/dbus/v5"
)

type bluetoothModule struct {
	receive     chan bool
	send        chan modules.Event
	device      string
	connected   bool
	powered     bool
	channel     chan *dbus.Signal
	conn        *dbus.Conn
	opts        Options
	initialOpts Options
}

func (mod *bluetoothModule) Dependencies() []string {
	return []string{}
}

func (mod *bluetoothModule) Name() string {
	return "bluetooth"
}

func New() modules.Module {
	return &bluetoothModule{}
}

func (mod *bluetoothModule) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *bluetoothModule) initState() error {
	mgr := mod.conn.Object("org.bluez", dbus.ObjectPath("/"))
	var objs map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	if err := mgr.Call(
		"org.freedesktop.DBus.ObjectManager.GetManagedObjects",
		0,
	).Store(&objs); err != nil {
		return fmt.Errorf("GetManagedObjects failed: %v", err)
	}

	var gotAdapter, gotDevice bool

	for _, ifaces := range objs {
		if props, ok := ifaces["org.bluez.Adapter1"]; ok && !gotAdapter {
			if v, exists := props["Powered"]; exists {
				mod.powered = v.Value().(bool)
				gotAdapter = true
			}
		}
		if props, ok := ifaces["org.bluez.Device1"]; ok && !gotDevice {
			if v, exists := props["Connected"]; exists {
				mod.connected = v.Value().(bool)
			}
			if v, exists := props["Name"]; exists {
				mod.device = v.Value().(string)
			}
			gotDevice = true
		}
		if gotAdapter && gotDevice {
			break
		}
	}

	if !gotAdapter {
		return fmt.Errorf("no org.bluez.Adapter1 found in managed objects")
	}
	return nil
}

func (mod *bluetoothModule) setConnection() error {
	conn, err := dbus.SystemBus()
	if err != nil {
		return fmt.Errorf("Failed to connect to system bus: %v", err)
	}
	rule := "type='signal',sender='org.bluez',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged'"
	call := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, rule)
	if call.Err != nil {
		return fmt.Errorf("Failed to add match rule: %v", call.Err)
	}
	ch := make(chan *dbus.Signal, 10)
	conn.Signal(ch)
	mod.conn = conn
	mod.channel = ch
	return nil
}

func (mod *bluetoothModule) checkActivity(signal *dbus.Signal) error {
	if len(signal.Body) < 3 {
		return fmt.Errorf("invalid signal")
	}

	iface, ok := signal.Body[0].(string)
	if !ok {
		return fmt.Errorf("invalid interface value")
	}

	changedProps, ok := signal.Body[1].(map[string]dbus.Variant)
	if !ok {
		return fmt.Errorf("value not the expected map")
	}

	path := signal.Path

	switch iface {
	case "org.bluez.Device1":
		if val, exists := changedProps["Connected"]; exists {
			mod.connected = val.Value().(bool)

			obj := mod.conn.Object("org.bluez", path)
			nameVal, err := obj.GetProperty("org.bluez.Device1.Name")
			if err != nil {
				return fmt.Errorf("Failed to get device name for %s: %v", path, err)
			}
			mod.device = nameVal.Value().(string)
		}

	case "org.bluez.Adapter1":
		if val, exists := changedProps["Powered"]; exists {
			mod.powered = val.Value().(bool)
		}
	}

	return nil
}

func (mod *bluetoothModule) Run() (<-chan bool, chan<- modules.Event, error) {
	err := mod.setConnection()
	if err != nil {
		return nil, nil, err
	}

	mod.send = make(chan modules.Event)
	mod.receive = make(chan bool)

	mod.initialOpts = mod.opts
	err = mod.initState()
	if err != nil {
		return nil, nil, err
	}

	go func() {
		for {
			select {

			case sig := <-mod.channel:
				err = mod.checkActivity(sig)
				if err != nil {
					continue
				}
				mod.receive <- true

			case e := <-mod.send:
				switch ev := e.VaxisEvent.(type) {
				case vaxis.Mouse:
					if ev.EventType != vaxis.EventPress {
						break
					}
					btn := config.ButtonName(ev)
					if mod.opts.OnClick.Dispatch(btn, &mod.initialOpts, &mod.opts) {
						mod.receive <- true
					}

				case modules.FocusIn:
					if mod.opts.OnClick.HoverIn(&mod.opts) {
						mod.receive <- true
					}

				case modules.FocusOut:
					if mod.opts.OnClick.HoverOut(&mod.opts) {
						mod.receive <- true
					}
				}
			}
		}
	}()

	return mod.receive, mod.send, nil
}

func (mod *bluetoothModule) Render() []modules.EventCell {
	style := vaxis.Style{
		Foreground: mod.opts.Fg.Go(),
		Background: mod.opts.Bg.Go(),
	}

	data := struct {
		Device string
	}{}

	format := mod.opts.Format

	if !mod.powered {
		format = mod.opts.NoConnection.Format
		style.Foreground = mod.opts.NoConnection.Fg.Go()
	}

	if mod.connected {
		data.Device = mod.device
	}

	var buf bytes.Buffer
	if err := format.Execute(&buf, data); err != nil {
		return nil
	}

	chars := vaxis.Characters(buf.String())
	r := make([]modules.EventCell, len(chars))
	for i, ch := range chars {
		r[i] = modules.EventCell{
			C: vaxis.Cell{
				Character: ch,
				Style:     style,
			},
			Mod:        mod,
			MouseShape: mod.opts.Cursor.Go(),
		}
	}
	return r
}
