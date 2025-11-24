// Copyright (c) 2025 Nekorg All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// SPDX-License-Identifier: bsd

package units

import (
	"fmt"
)

const (
	Byte float64 = 1

	// IEC
	KiB = 1024 * Byte
	MiB = 1024 * KiB
	GiB = 1024 * MiB
	TiB = 1024 * GiB

	// SI (for those people...)
	KB = 1000 * Byte
	MB = 1000 * KB
	GB = 1000 * MB
	TB = 1000 * GB
)

// I don't think anyone with more than a PB/PiB of
// disk space cares for this, so it is what it is

type Unit struct {
	Div  float64
	Name string
}

type System int

const (
	IEC System = iota
	SI
)

func ParseSystem(s string) System {
	switch s {
	case "si", "SI", "metric", "decimal":
		return SI
	default:
		return IEC
	}
}

var (
	unitsIEC = []Unit{
		{TiB, "TiB"},
		{GiB, "GiB"},
		{MiB, "MiB"},
		{KiB, "KiB"},
		{Byte, "B"},
	}

	unitsSI = []Unit{
		{TB, "TB"},
		{GB, "GB"},
		{MB, "MB"},
		{KB, "KB"},
		{Byte, "B"},
	}
)

var parseTable = map[string]Unit{
	"b": {Byte, "B"}, "byte": {Byte, "B"}, "bytes": {Byte, "B"},
	"kib": {KiB, "KiB"}, "kibibyte": {KiB, "KiB"}, "kibibytes": {KiB, "KiB"},
	"mib": {MiB, "MiB"}, "mibibyte": {MiB, "MiB"}, "mibibytes": {MiB, "MiB"},
	"gib": {GiB, "GiB"}, "gibibyte": {GiB, "GiB"}, "gibibytes": {GiB, "GiB"},
	"tib": {TiB, "TiB"}, "tebibyte": {TiB, "TiB"}, "tebibytes": {TiB, "TiB"},

	"kb": {KB, "KB"}, "kilobyte": {KB, "KB"}, "kilobytes": {KB, "KB"},
	"mb": {MB, "MB"}, "megabyte": {MB, "MB"}, "megabytes": {MB, "MB"},
	"gb": {GB, "GB"}, "gigabyte": {GB, "GB"}, "gigabytes": {GB, "GB"},
	"tb": {TB, "TB"}, "terabyte": {TB, "TB"}, "terabytes": {TB, "TB"},
}

func table(sys System) []Unit {
	switch sys {
	case SI:
		return unitsSI
	default:
		return unitsIEC
	}
}

func Choose(bytes uint64, sys System) Unit {
	units := table(sys)
	for _, u := range units {
		if float64(bytes)/u.Div >= 1.0 {
			return u
		}
	}

	return units[len(units)-1]
}

func Format(v uint64, u Unit) float64 {
	return float64(v) / u.Div
}

func ParseUnit(s string) (Unit, error) {
	u, ok := parseTable[s]
	if !ok {
		return Unit{}, fmt.Errorf("cannot parse unit name %q", s)
	}

	return u, nil
}
