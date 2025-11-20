package main

import (
	"flag"

	"github.com/nekorg/pawbar/pkg/dbusmenukitty"
)

func main() {
	var x, y int

	// flag.StringVar(&service, "service", "", "DBus service name exposing a dbusmenu (e.g. org.freedesktop.network-manager-applet)")
	flag.IntVar(&x, "x", 0, "X coordinate for panel (pixels)")
	flag.IntVar(&y, "y", 0, "Y coordinate for panel (pixels)")
	flag.Parse()

	// LaunchMenu will not return until the panel closes (or an error occurs).
	dbusmenukitty.LaunchMenu(x, y)
}
