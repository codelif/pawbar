// Copyright (c) 2025 Nekorg All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// SPDX-License-Identifier: bsd

package menu

func MaxLengthLabel(labels []Item) int {
	if len(labels) == 0 {
		return 0
	}
  
  maxLen := 0
	for _, l := range labels {
		curLen := len(l.Label.Display)

		if l.IconData != nil || l.IconName != "" {
			curLen += 2
		}
		if curLen > maxLen {
			maxLen = curLen
		}
	}

	return maxLen
}
