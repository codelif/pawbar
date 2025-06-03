package mpris

import (
	"bytes"
	"fmt"
	"strings"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/godbus/dbus/v5"
)

type MprisModule struct {
	receive     chan bool
	send        chan modules.Event
	format      Format
	opts        Options
	initialOpts Options
	conn        *dbus.Conn
	channel     chan *dbus.Signal
	artists     []string
	title       string
}

func New() modules.Module { return &MprisModule{} }

func (mod *MprisModule) Name() string                                  { return "mpris" }
func (mod *MprisModule) Dependencies() []string                        { return nil }
func (mod *MprisModule) Channels() (<-chan bool, chan<- modules.Event) { return mod.receive, mod.send }

type Format int

func (f *Format) toggle() { *f ^= 1 }

const (
	FormatPause Format = iota
	FormatPlay
)

func (mod *MprisModule) Connection() error {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return err
	}

	call := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		"type='signal',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged'")
	if call.Err != nil {
		return err
	}
	ch := make(chan *dbus.Signal, 10)
	conn.Signal(ch)
	mod.conn = conn
	mod.channel = ch
	return nil
}

func (mod *MprisModule) CatchEvent(signal *dbus.Signal) error {
	if len(signal.Body) < 3 {
		return fmt.Errorf("Distorted Signals")
	}

	iface, ok := signal.Body[0].(string)
	if !ok || iface != "org.mpris.MediaPlayer2.Player" {
		return fmt.Errorf("Invalid Interface found")
	}

	changedProps, ok := signal.Body[1].(map[string]dbus.Variant)
	if !ok {
		return fmt.Errorf("Error in parsing changed properties")
	}
	for prop, val := range changedProps {
		switch prop {
		case "PlaybackStatus":
			if status, ok := val.Value().(string); ok {
				if status == "Playing" {
					mod.format = FormatPlay
				} else if status == "Paused" {
					mod.format = FormatPause
				}
			}

		case "Metadata":
			metaMap, ok := val.Value().(map[string]dbus.Variant)
			if !ok {
				break
			}

			if titleVar, found := metaMap["xesam:title"]; found {
				if title, ok := titleVar.Value().(string); ok {
					mod.title = title
				}
			}

			if artistVar, found := metaMap["xesam:artist"]; found {
				if artists, ok := artistVar.Value().([]string); ok {
					if len(artists) > 0 {
						mod.artists = artists
					}
				}
			}
		}
	}
	return nil
}

func (mod *MprisModule) SendEvent() error {
	var busNames []string
	err := mod.conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&busNames)
	if err != nil {
		return fmt.Errorf("failed to list bus names: %w", err)
	}

	var activePlayer string
	for _, name := range busNames {
		if strings.HasPrefix(name, "org.mpris.MediaPlayer2.") {
			obj := mod.conn.Object(name, dbus.ObjectPath("/org/mpris/MediaPlayer2"))

			variant, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.PlaybackStatus")
			if err == nil {
				status, ok := variant.Value().(string)
				if ok && (status == "Playing" || status == "Paused") {
					activePlayer = name
					break
				}
			} else {
				// fallback to this player if no better match is found later
				if activePlayer == "" {
					activePlayer = name
				}
			}
		}
	}

	if activePlayer == "" {
		return fmt.Errorf("no MPRIS player found on the session bus")
	}

	obj := mod.conn.Object(activePlayer, dbus.ObjectPath("/org/mpris/MediaPlayer2"))
	call := obj.Call("org.mpris.MediaPlayer2.Player.PlayPause", 0)
	if call.Err != nil {
		return fmt.Errorf("failed to send PlayPause to %s: %w", activePlayer, call.Err)
	}

	return nil
}

func (mod *MprisModule) Run() (<-chan bool, chan<- modules.Event, error) {
	err := mod.Connection()
	if err != nil {
		return nil, nil, err
	}

	mod.send = make(chan modules.Event)
	mod.receive = make(chan bool)

	mod.initialOpts = mod.opts
	if err != nil {
		return nil, nil, err
	}

	go func() {
		for {
			select {

			case sig := <-mod.channel:
				err = mod.CatchEvent(sig)
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
					if btn == "left" {
						err = mod.SendEvent()
						if err != nil {
							continue
						}
						mod.format.toggle()
						mod.receive <- true
					}
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

func (mod *MprisModule) Render() []modules.EventCell {
	style := vaxis.Style{
		Foreground: mod.opts.Fg.Go(),
		Background: mod.opts.Bg.Go(),
	}

	artists := strings.Join(mod.artists, ",")
	data := struct {
		Icon    string
		Artists string
		Title   string
	}{}

	var tpl config.Format

	switch mod.format {
	case FormatPlay:
		data.Icon = string(mod.opts.Play.Icon)
		data.Artists = artists
		data.Title = mod.title
		tpl = mod.opts.Play.Format
	case FormatPause:
		data.Icon = string(mod.opts.Pause.Icon)
		data.Artists = artists
		data.Title = mod.title
		tpl = mod.opts.Pause.Format
	default:
		tpl = mod.opts.Format
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
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
