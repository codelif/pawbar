/*
 * Copyright (c) 2025 Nekorg All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 *
 * SPDX-License-Identifier: bsd
 */

#include <GLFW/glfw3.h>
#include <stdlib.h>

typedef struct {
  int width;
  int height;
  int refreshRate;

  float scaleX;
  float scaleY;
} monitor;

monitor *get_monitor_info();
void free_monitor(monitor *m);
