package wifi

import (
	"bytes"
	"fmt"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	nm "github.com/Wifx/gonetworkmanager/v3"
	"github.com/nekorg/pawbar/internal/config"
	"github.com/nekorg/pawbar/internal/modules"
	"github.com/nekorg/pawbar/internal/utils"
	"github.com/godbus/dbus/v5"
)

const deviceIndex = 2

type wifiModule struct {
	SSID          string
	InterfaceName string
	accessPoint   nm.AccessPoint
	receive       chan bool
	send          chan modules.Event
	nmgr          nm.NetworkManager

	opts        Options
	initialOpts Options

	currentTickerInterval time.Duration
	ticker                *time.Ticker
}

func (mod *wifiModule) Dependencies() []string {
	return []string{}
}

func (mod *wifiModule) Name() string {
	return "wifi"
}

func New() modules.Module {
	return &wifiModule{}
}

func (mod *wifiModule) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *wifiModule) GetAccessPoint(devicePath dbus.ObjectPath, sig dbus.Signal) error {
	if sig.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" && sig.Path == devicePath {

		interfaceName := sig.Body[0].(string)
		changes := sig.Body[1].(map[string]dbus.Variant)

		if interfaceName == nm.DeviceWirelessInterface {
			if v, ok := changes["ActiveAccessPoint"]; ok {
				newPath, ok := v.Value().(dbus.ObjectPath)
				if !ok {
					return fmt.Errorf("ActiveAccessPoint variant is %T, not ObjectPath", v.Value())
				}
				if newPath == "/" {
					mod.SSID = ""
					mod.accessPoint = nil
					mod.InterfaceName = interfaceName
					return nil
				}
				err := mod.Connection(devicePath)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (mod *wifiModule) Connection(devicePath dbus.ObjectPath) error {
	ssid := ""
	Interface := ""
	var ap nm.AccessPoint

	wdev, err := nm.NewDeviceWireless(devicePath)
	if err != nil {
		return err
	}

	Interface, err = wdev.GetPropertyInterface()
	if err != nil {
		return err
	}
	ap, err = wdev.GetPropertyActiveAccessPoint()
	if err != nil {
		return err
	}

	if ap.GetPath() != "/" {
		ssid, err = ap.GetPropertySSID()
		if err != nil {
			return err
		}
	}

	mod.accessPoint = ap
	mod.SSID = ssid
	mod.InterfaceName = Interface
	return nil
}

func (mod *wifiModule) Subscribe() (dbus.ObjectPath, <-chan *dbus.Signal, error) {
	nmgr, err := nm.NewNetworkManager()
	if err != nil {
		return "", nil, err
	}
	mod.nmgr = nmgr
	devicePath := dbus.ObjectPath(
		fmt.Sprintf("/org/freedesktop/NetworkManager/Devices/%d", deviceIndex),
	)
	sigs := nmgr.Subscribe()

	return devicePath, sigs, nil
}

func (mod *wifiModule) GetStrenght(ap nm.AccessPoint) (int, error) {
	if ap.GetPath() == "/" {
		return 0, nil
	}
	strength, err := ap.GetPropertyStrength()
	if err != nil {
		return 0, err
	}
	return int(strength), nil
}

func (mod *wifiModule) Run() (<-chan bool, chan<- modules.Event, error) {
	devicePath, sigs, err := mod.Subscribe()
	if err != nil {
		return nil, nil, err
	}
	mod.send = make(chan modules.Event)
	mod.receive = make(chan bool)

	mod.initialOpts = mod.opts

	err = mod.Connection(devicePath)
	if err != nil {
		return nil, nil, err
	}

	go func() {
		defer mod.nmgr.Unsubscribe()
		mod.currentTickerInterval = mod.opts.Tick.Go()
		mod.ticker = time.NewTicker(mod.currentTickerInterval)
		defer mod.ticker.Stop()
		for {
			select {

			case <-mod.ticker.C:
				mod.receive <- true

			case sig := <-sigs:
				err := mod.GetAccessPoint(devicePath, *sig)
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
					mod.ensureTickInterval()

				case modules.FocusIn:
					if mod.opts.OnClick.HoverIn(&mod.opts) {
						mod.receive <- true
					}
					mod.ensureTickInterval()

				case modules.FocusOut:
					if mod.opts.OnClick.HoverOut(&mod.opts) {
						mod.receive <- true
					}
					mod.ensureTickInterval()
				}
			}
		}
	}()

	return mod.receive, mod.send, nil
}

func (mod *wifiModule) ensureTickInterval() {
	if mod.opts.Tick.Go() != mod.currentTickerInterval {
		mod.currentTickerInterval = mod.opts.Tick.Go()
		mod.ticker.Reset(mod.currentTickerInterval)
	}
}

func (mod *wifiModule) Render() []modules.EventCell {
	style := vaxis.Style{
		Foreground: mod.opts.Fg.Go(),
		Background: mod.opts.Bg.Go(),
	}

	data := struct {
		Icon      string
		SSID      string
		Interface string
	}{}

	format := mod.opts.Format

	if mod.SSID == "" {
		format = mod.opts.NoConnection.Format
		style.Foreground = mod.opts.NoConnection.Fg.Go()
	} else {
		strength, err := mod.GetStrenght(mod.accessPoint)
		if err == nil {
			idx := utils.Clamp((len(mod.opts.Icons)-1)*strength/100, 0, len(mod.opts.Icons)-1)
			data.Icon = string(mod.opts.Icons[idx])
		}
		data.SSID = mod.SSID
		data.Interface = mod.InterfaceName
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
