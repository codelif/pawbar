#!/usr/bin/bash
# Copyright (c) 2025 Nekorg All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.
#
# SPDX-License-Identifier: bsd


sudo cp pawbar /usr/local/bin/
mkdir -p "$HOME/.config/pawbar/"
[ ! -f "$HOME/.config/pawbar/pawbar.yaml" ] && echo -e  "right:\n  - battery\n  - sep\n  - clock" > "$HOME/.config/pawbar/pawbar.yaml"
