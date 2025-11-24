/*
 * Copyright (c) 2025 Nekorg All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 *
 * SPDX-License-Identifier: bsd
 */

#include "monitor.h"

monitor *get_monitor_info() {
  if (!glfwInit()) {
    return NULL;
  }

  GLFWmonitor *primaryMonitor = glfwGetPrimaryMonitor();
  if (!primaryMonitor) {
    glfwTerminate();
    return NULL;
  }

  const GLFWvidmode *mode = glfwGetVideoMode(primaryMonitor);
  if (!mode) {
    glfwTerminate();
    return NULL;
  }

  float x, y;
  glfwGetMonitorContentScale(primaryMonitor, &x, &y);

  monitor *m = (monitor *)malloc(sizeof(monitor));
  m->height = mode->height;
  m->width = mode->width;
  m->refreshRate = mode->refreshRate;
  m->scaleX = x;
  m->scaleY = y;

  glfwTerminate();
  return m;
}

void free_monitor(monitor *m) { free(m); }
