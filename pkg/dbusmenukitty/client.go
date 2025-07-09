package dbusmenukitty

import (
	"fmt"
	"io"
	"log"
	"time"
	"strings"

	"github.com/codelif/katnip"
	"github.com/fxamacker/cbor/v2"
	"github.com/godbus/dbus/v5"
)

type Layout struct {
	Id         int32
	Properties map[string]dbus.Variant
	Layouts    []Layout
}

func LaunchMenu(x, y int) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		log.Fatalf("error connecting to session bus")
	}
	defer conn.Close()
	obj := conn.Object("org.freedesktop.network-manager-applet", "/org/ayatana/NotificationItem/nm_applet/Menu")
	// obj = conn.Object(":1.25", "/org/ayatana/NotificationItem/nm_applet/Menu")

	var s string
	obj.StoreProperty("com.canonical.dbusmenu.Status", &s)

	fmt.Printf("Status: %s\n", s)

	c := obj.Call("com.canonical.dbusmenu.GetLayout", 0, 0, -1, []string{})
	var revision uint32
	var layout Layout
	c.Store(&revision, &layout)
	fmt.Printf("Revision: %v\n", revision)
	// printLayout(layout, 0)

	kn := katnip.NewPanel("leaf", katnip.Config{
    Position: katnip.Vector{X: x, Y: y},
		Size:  katnip.Vector{X: 40, Y: 20},
		Edge:  katnip.EdgeNone,
		Layer: katnip.LayerTop,
	})

	kn.Start()

	w := kn.Writer()
	e := cbor.NewEncoder(w)
	serialiseLayout(e, layout)

	kn.Wait()
}

func printLayout(l Layout, indent int) {
	if _, ok := l.Properties["icon-data"]; ok {
		l.Properties["icon-data"] = dbus.MakeVariant("omitted...")
	}
	fmt.Printf("%sId: %v\n%sProperties:%v\n%sLayout:\n\n", strings.Repeat(" ", indent*4), l.Id, strings.Repeat(" ", indent*4), l.Properties, strings.Repeat(" ", indent*4))
	for _, li := range l.Layouts {
		printLayout(li, indent+1)
	}
}

type Label struct {
	Display     string
	AccessKey   rune
	AccessIndex int
	Found       bool
}

func ParseAccessKey(label string) Label {
	runes := []rune(label)
	n := len(runes)

	var output []rune
	outPos := 0
	var result Label

	for i := 0; i < n; {
		if runes[i] == '_' {
			if i+1 < n && runes[i+1] == '_' {
				output = append(output, '_')
				outPos++
				i += 2
			} else {
				if !result.Found && i+1 < n {
					result.Found = true
					result.AccessKey = runes[i+1]
					result.AccessIndex = outPos
					output = append(output, runes[i+1])
					outPos++
					i += 2
				} else {
					i++
				}
			}
		} else {
			output = append(output, runes[i])
			outPos++
			i++
		}
	}

	result.Display = string(output)
	return result
}

type JLay struct {
	Blocks []Label
}

func serialiseLayout(enc *cbor.Encoder, l Layout) {
	j := JLay{}
	for _, l := range l.Layouts {
		labelV, ok := l.Properties["label"]
		if ok {
			label := labelV.Value().(string)
			j.Blocks = append(j.Blocks, ParseAccessKey(label))
		}
	}

	enc.Encode(j)
}

func init() {
	katnip.RegisterFunc("leaf", leaf)
}

func leaf(k *katnip.Kitty, rw io.ReadWriter) int {
	_ = cbor.NewEncoder(rw)
	dec := cbor.NewDecoder(rw)

	var l JLay
	dec.Decode(&l)
	for _, v := range l.Blocks {
		fmt.Println(v.Display)
	}
	time.Sleep(10 * time.Second)
	return 0
}
