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
