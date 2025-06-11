package dbusmenukitty

import (
	"fmt"
	"log"
	"strings"

	"github.com/godbus/dbus/v5"
)

type Layout struct {
	Id         int32
	Properties map[string]dbus.Variant
	Layouts    []Layout
}

func Onit() {
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
	printLayout(layout, 0)
}

func printLayout(l Layout, indent int) {
  if _,ok:=l.Properties["icon-data"]; ok {
    l.Properties["icon-data"] = dbus.MakeVariant("omitted...")
  }
	fmt.Printf("%sId: %v\n%sProperties:%v\n%sLayout:\n\n", strings.Repeat(" ", indent*4), l.Id, strings.Repeat(" ", indent*4), l.Properties, strings.Repeat(" ", indent*4))
	for _, li := range l.Layouts {
		printLayout(li, indent+1)
	}
}
