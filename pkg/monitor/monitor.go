// Copyright (c) 2025 Nekorg All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// SPDX-License-Identifier: bsd

package monitor

// #cgo LDFLAGS: -lglfw
// #include "monitor.h"
import "C"
import "fmt"

type Monitor struct {
	Width, Height, RefreshRate int
	Scale                      Scale
}

type Scale struct {
	X, Y float64
}

func GetMonitorInfo() (Monitor, error) {
	m := C.get_monitor_info()

	if m == nil {
		return Monitor{}, fmt.Errorf("failed get monitor info")
	}

	mon := Monitor{
		Width:       int(m.width),
		Height:      int(m.height),
		RefreshRate: int(m.refreshRate),
		Scale: Scale{
			X: float64(m.scaleX),
			Y: float64(m.scaleY),
		},
	}

	C.free_monitor(m)

	return mon, nil
}
