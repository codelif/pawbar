// Copyright (c) 2025 Nekorg All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// SPDX-License-Identifier: bsd

package cpu

import (
	"bytes"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/nekorg/pawbar/internal/config"
	"github.com/nekorg/pawbar/internal/modules"
	"github.com/shirou/gopsutil/v3/cpu"
)

type CpuModule struct {
	receive chan bool
	send    chan modules.Event

	opts        Options
	initialOpts Options

	highStart     time.Time
	highTriggered bool

	currentTickerInterval time.Duration
	ticker                *time.Ticker
}

func (mod *CpuModule) Dependencies() []string {
	return nil
}

func (mod *CpuModule) Run() (<-chan bool, chan<- modules.Event, error) {
	mod.receive = make(chan bool)
	mod.send = make(chan modules.Event)
	mod.initialOpts = mod.opts

	go func() {
		mod.currentTickerInterval = mod.opts.Tick.Go()
		mod.ticker = time.NewTicker(mod.currentTickerInterval)
		defer mod.ticker.Stop()
		for {
			select {
			case <-mod.ticker.C:
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

func (mod *CpuModule) ensureTickInterval() {
	if mod.opts.Tick.Go() != mod.currentTickerInterval {
		mod.currentTickerInterval = mod.opts.Tick.Go()
		mod.ticker.Reset(mod.currentTickerInterval)
	}
}

func (mod *CpuModule) Render() []modules.EventCell {
	percent, err := cpu.Percent(0, false)
	if err != nil || len(percent) == 0 {
		return nil
	}
	usage := int(percent[0])

	threshold := mod.opts.Threshold.Percent.Go()
	if usage > threshold {
		if mod.highStart.IsZero() {
			mod.highStart = time.Now()
		} else if !mod.highTriggered && time.Since(mod.highStart) >= mod.opts.Threshold.For.Go() {
			mod.highTriggered = true
		}
	} else {
		mod.highStart = time.Time{}
		mod.highTriggered = false
	}

	style := vaxis.Style{}
	if mod.highTriggered {
		style.Foreground = mod.opts.Threshold.Fg.Go()
		style.Background = mod.opts.Threshold.Bg.Go()

	} else {
		style.Foreground = mod.opts.Fg.Go()
		style.Background = mod.opts.Bg.Go()
	}

	var buf bytes.Buffer
	_ = mod.opts.Format.Execute(&buf, struct{ Percent int }{usage})

	rch := vaxis.Characters(buf.String())
	r := make([]modules.EventCell, len(rch))

	for i, ch := range rch {
		r[i] = modules.EventCell{C: vaxis.Cell{Character: ch, Style: style}, Mod: mod, MouseShape: mod.opts.Cursor.Go()}
	}
	return r
}

func (mod *CpuModule) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *CpuModule) Name() string {
	return "cpu"
}
