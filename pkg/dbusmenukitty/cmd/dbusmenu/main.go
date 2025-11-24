// Copyright (c) 2025 Nekorg All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// SPDX-License-Identifier: bsd

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
