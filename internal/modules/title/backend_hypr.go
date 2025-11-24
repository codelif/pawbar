// Copyright (c) 2025 Nekorg All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// SPDX-License-Identifier: bsd

package title

import (
	"strings"

	"github.com/nekorg/pawbar/internal/services/hypr"
)

type hyprBackend struct {
	svc   *hypr.Service
	ev    chan hypr.HyprEvent
	class string
	title string
	sig   chan struct{}
}

func newHyprBackend(s *hypr.Service) backend {
	b := &hyprBackend{
		svc: s,
		ev:  make(chan hypr.HyprEvent),
		sig: make(chan struct{}, 1),
	}

	activews := hypr.GetActiveWorkspace()
	clients := hypr.GetClients()

	b.class = ""
	for _, c := range clients {
		if c.Address == activews.Lastwindow {
			b.class = c.Class
		}
	}

	b.title = hypr.GetActiveWorkspace().Lastwindowtitle
	b.svc.RegisterChannel("activewindow", b.ev)
	go b.loop()
	return b
}

func (b *hyprBackend) loop() {
	for e := range b.ev {
		b.class, b.title, _ = strings.Cut(e.Data, ",")
		b.signal()
	}
}

func (b *hyprBackend) signal() {
	select {
	case b.sig <- struct{}{}:
	default:
	}
}

func (b *hyprBackend) Window() Window {
	return Window{Title: b.title, Class: b.class}
}
func (b *hyprBackend) Events() <-chan struct{} { return b.sig }
