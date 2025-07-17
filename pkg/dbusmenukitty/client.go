package dbusmenukitty

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/codelif/katnip"
	"github.com/codelif/pawbar/pkg/dbusmenukitty/menu"
	"github.com/codelif/pawbar/pkg/dbusmenukitty/tui"
	"github.com/codelif/xdgicons"
	"github.com/fxamacker/cbor/v2"
	"github.com/godbus/dbus/v5"
)

var iconLookup = xdgicons.NewIconLookupWithConfig(xdgicons.LookupConfig{FallbackTheme: "Adwaita"})

type Layout struct {
	Id         int32
	Properties map[string]dbus.Variant
	Children   []Layout
}

func LaunchMenu(x, y int) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		log.Fatalf("error connecting to session bus")
	}
	defer conn.Close()

	busname := "org.freedesktop.network-manager-applet"
	path := "/org/ayatana/NotificationItem/nm_applet/Menu"

	busname = "org.blueman.Tray"
	path = "/org/blueman/sni/menu"

	obj := conn.Object(busname, dbus.ObjectPath(path))
	// obj = conn.Object(":1.25", "/org/ayatana/NotificationItem/nm_applet/Menu")

	var s string
	obj.StoreProperty("com.canonical.dbusmenu.Status", &s)

	fmt.Printf("Status: %s\n", s)

	c := obj.Call("com.canonical.dbusmenu.GetLayout", 0, 0, -1, []string{})
	var revision uint32
	var layout Layout
	c.Store(&revision, &layout)
	fmt.Printf("Revision: %v\n", revision)
	printLayout(layout, 0)
	// fmt.Printf("%+v\n", FlattenLayout(layout))
	CreatePanel(x, y, FlattenLayout(layout))
}

func printLayout(l Layout, indent int) {
	if _, ok := l.Properties["icon-data"]; ok {
		l.Properties["icon-data"] = dbus.MakeVariant("omitted...")
	}
	fmt.Printf("%sId: %v\n%sProperties:%v\n%sLayout:\n\n", strings.Repeat(" ", indent*4), l.Id, strings.Repeat(" ", indent*4), l.Properties, strings.Repeat(" ", indent*4))
	for _, li := range l.Children {
		printLayout(li, indent+1)
	}
}

func FlattenLayout(parent Layout) (items []menu.Item) {
	for _, layout := range parent.Children {
		item := menu.Item{Id: layout.Id}
		if itemType, ok := layout.Properties["type"]; ok {
			item.Type = itemType.Value().(string)
		} else {
			item.Type = menu.ItemStandard
		}
		if label, ok := layout.Properties["label"]; ok {
			item.Label = menu.ParseLabel(label.Value().(string))
		}
		if enabled, ok := layout.Properties["enabled"]; ok {
			item.Enabled = enabled.Value().(bool)
		} else {
			item.Enabled = true
		}
		if visible, ok := layout.Properties["visible"]; ok {
			item.Visible = visible.Value().(bool)
		} else {
			item.Visible = true
		}
		if iconName, ok := layout.Properties["icon-name"]; ok {
			item.IconName = iconName.Value().(string)
			var icon xdgicons.Icon
			if strings.HasSuffix(item.IconName, "-symbolic") {
				icon, _ = iconLookup.Lookup(item.IconName)
			} else {
				icon, _ = iconLookup.FindBestIcon([]string{item.IconName + "-symbolic", item.IconName}, 48, 1)
			}
			item.Icon = icon
      fmt.Println(icon, item.IconName)
		}
		if iconData, ok := layout.Properties["icon-data"]; ok {
			item.IconData = iconData.Value().([]byte)
		}
		if shortcut, ok := layout.Properties["shortcut"]; ok {
			item.Shortcut = shortcut.Value().([][]string)
		}
		if toggleType, ok := layout.Properties["toggle-type"]; ok {
			item.ToggleType = toggleType.Value().(string)
		}
		if toggleState, ok := layout.Properties["toggle-state"]; ok {
			item.ToggleState = toggleState.Value().(int32)
		} else {
			item.ToggleState = -1
		}
		if childrenDisplay, ok := layout.Properties["children-display"]; ok {
			childrenDisplay := childrenDisplay.Value().(string)
			item.HasChildren = childrenDisplay == "submenu"
		}

		items = append(items, item)
	}

	return items
}

func init() {
	katnip.RegisterFunc("leaf", tui.Leaf)
}

func CreatePanel(x, y int, MenuItems []menu.Item) *katnip.Panel {
	conf := katnip.Config{
		Position:    katnip.Vector{X: x, Y: y},
		Size:        katnip.Vector{X: 1, Y: 1},
		Edge:        katnip.EdgeNone,
		Layer:       katnip.LayerTop,
		FocusPolicy: katnip.FocusOnDemand,
		KittyOverrides: []string{
			"font_size=16",
			"cursor_trail=0",
			"paste_actions=replace-dangerous-control-codes",
			"map kitty_mod+equal       no_op",
			"map kitty_mod+plus        no_op",
			"map kitty_mod+kp_add      no_op",
			"map cmd+plus              no_op",
			"map cmd+equal             no_op",
			"map shift+cmd+equal       no_op",
			"map kitty_mod+minus       no_op",
			"map kitty_mod+kp_subtract no_op",
			"map cmd+minus             no_op",
			"map shift+cmd+minus       no_op",
			"map kitty_mod+backspace   no_op",
			"map cmd+0                 no_op",
		},
	}

	kn := katnip.NewPanel("leaf", conf)
	kn.Start()
	enc := cbor.NewEncoder(kn.Writer())
	enc.Encode(MenuItems)

	io.Copy(os.Stdout, kn.Reader())
	kn.Wait()
	return kn
}
