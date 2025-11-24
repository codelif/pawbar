// Copyright (c) 2025 Nekorg All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// SPDX-License-Identifier: bsd

package tray

import (
	"bytes"
	"strconv"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/nekorg/pawbar/internal/config"
	"github.com/nekorg/pawbar/internal/modules"
	"github.com/nekorg/pawbar/internal/services/sni"
	"github.com/nekorg/pawbar/pkg/dbusmenukitty"
	"gopkg.in/yaml.v3"
)

func init() {
	config.Register("tray", func(n *yaml.Node) (modules.Module, error) {
		// no options yet
		return &Module{}, nil
	})
}

type Module struct {
	svc      *sni.Service
	receive  chan bool
	send     chan modules.Event
	lastList []sni.Item
}

func (m *Module) Name() string                                  { return "tray" }
func (m *Module) Dependencies() []string                        { return []string{"sni"} }
func (m *Module) Channels() (<-chan bool, chan<- modules.Event) { return m.receive, m.send }

func (m *Module) Run() (<-chan bool, chan<- modules.Event, error) {
	svc, ok := sni.Register()
	if !ok {
		return nil, nil, nil
	}
	m.svc = svc
	m.receive = make(chan bool, 4)
	m.send = make(chan modules.Event, 8)

	// Subscribe to SNI updates
	evs := svc.IssueListener()
	m.lastList = svc.Items()

	go func() {
		for {
			select {
			case <-evs:
				m.lastList = svc.Items()
				m.receive <- true
			case e := <-m.send:
				switch ev := e.VaxisEvent.(type) {
				case vaxis.Mouse:
					if ev.EventType != vaxis.EventPress {
						break
					}
					// Which “cell” (item) did we click? We encode metadata = index
					idx, _ := strconv.Atoi(e.Cell.Metadata)
					if idx < 0 || idx >= len(m.lastList) {
						break
					}
					item := m.lastList[idx]
					switch ev.Button {
					case vaxis.MouseLeftButton:
						_ = svc.Activate(item, int32(ev.XPixel), int32(ev.YPixel))
					case vaxis.MouseMiddleButton:
						_ = svc.SecondaryActivate(item, int32(ev.XPixel), int32(ev.YPixel))
					case vaxis.MouseRightButton:
						// Prefer item.ContextMenu if it exists; if menu path is published, open our TUI popup
						if item.MenuPath != "" {
							// TODO: this assumes 2x scale, use pkg/monitor to determine correct scale.
							go dbusmenukitty.LaunchMenu(ev.XPixel/2, ev.YPixel/2)
						} else {
							_ = svc.ContextMenu(item, int32(ev.XPixel), int32(ev.YPixel))
						}
					case vaxis.MouseWheelUp:
						_ = svc.Scroll(item, +120, "vertical")
					case vaxis.MouseWheelDown:
						_ = svc.Scroll(item, -120, "vertical")
					}
				}
			}
		}
	}()

	return m.receive, m.send, nil
}

func (m *Module) Render() []modules.EventCell {
	list := m.lastList
	if len(list) == 0 {
		return nil
	}
	// MVP text form: [icon-like]Title or Id
	// You can swap ’labelFor’ to prefer IconName, Title, Id, etc.
	labelFor := func(it sni.Item) string {
		if it.IconName != "" {
			return it.IconName // simple and compact
		}
		if it.Title != "" {
			return it.Title
		}
		if it.Id != "" {
			return it.Id
		}
		return "?"
	}

	var out []modules.EventCell
	style := vaxis.Style{} // inherit bar defaults
	for i, it := range list {
		// Add a space separator between items
		if i != 0 {
			out = append(out, modules.ECSPACE)
		}
		var buf bytes.Buffer
		buf.WriteString(labelFor(it))
		for _, ch := range vaxis.Characters(buf.String()) {
			out = append(out, modules.EventCell{
				C:        vaxis.Cell{Character: ch, Style: style},
				Metadata: strconv.Itoa(i), // index for click routing
				Mod:      m,
				// use pointer cursor to hint clickability
				MouseShape: vaxis.MouseShapeDefault,
			})
		}
	}
	return out
}
