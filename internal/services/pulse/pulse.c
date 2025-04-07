#include "pulse_wrapper.h"
#include <pulse/pulseaudio.h>
#include <pulse/thread-mainloop.h>
#include <pulse/volume.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static pa_threaded_mainloop *mloop = NULL;
static pa_context *context = NULL;

static void context_state_callback(pa_context *ctx, void *userdata) {
  pa_threaded_mainloop *m = (pa_threaded_mainloop *)userdata;
  pa_threaded_mainloop_signal(m, 0);
}

int pulse_init() {
  mloop = pa_threaded_mainloop_new();
  if (!mloop)
    return -1;
  pa_mainloop_api *mainloop_api = pa_threaded_mainloop_get_api(mloop);
  context = pa_context_new(mainloop_api, "GoPulseApp");
  if (!context) {
    pa_threaded_mainloop_free(mloop);
    return -1;
  }
  pa_context_set_state_callback(context, context_state_callback, mloop);
  if (pa_context_connect(context, NULL, PA_CONTEXT_NOFLAGS, NULL) < 0) {
    pa_context_unref(context);
    pa_threaded_mainloop_free(mloop);
    return -1;
  }
  pa_threaded_mainloop_lock(mloop);
  if (pa_threaded_mainloop_start(mloop) < 0) {
    pa_threaded_mainloop_unlock(mloop);
    return -1;
  }
  while (1) {
    pa_context_state_t state = pa_context_get_state(context);
    if (state == PA_CONTEXT_READY)
      break;
    if (!PA_CONTEXT_IS_GOOD(state)) {
      pa_threaded_mainloop_unlock(mloop);
      return -1;
    }
    pa_threaded_mainloop_wait(mloop);
  }
  pa_threaded_mainloop_unlock(mloop);
  return 0;
}

struct default_sink_data {
  char sink[256];
  int done;
};

static void server_info_callback(pa_context *c, const pa_server_info *i,
                                 void *userdata) {
  struct default_sink_data *data = (struct default_sink_data *)userdata;
  if (i && i->default_sink_name) {
    strncpy(data->sink, i->default_sink_name, sizeof(data->sink) - 1);
    data->sink[sizeof(data->sink) - 1] = '\0';
  } else {
    data->sink[0] = '\0';
  }
  data->done = 1;
  pa_threaded_mainloop_signal(mloop, 0);
}

int pulse_get_default_sink(char *sink, int sink_len) {
  if (!context || !mloop)
    return -1;
  pa_threaded_mainloop_lock(mloop);
  struct default_sink_data data;
  memset(&data, 0, sizeof(data));
  data.done = 0;
  pa_operation *op =
      pa_context_get_server_info(context, server_info_callback, &data);
  if (!op) {
    pa_threaded_mainloop_unlock(mloop);
    return -1;
  }
  while (!data.done)
    pa_threaded_mainloop_wait(mloop);
  pa_operation_unref(op);
  pa_threaded_mainloop_unlock(mloop);
  strncpy(sink, data.sink, sink_len - 1);
  sink[sink_len - 1] = '\0';
  return 0;
}

struct sink_info_data {
  double volume;
  int mute;
  int done;
};

static void sink_info_callback(pa_context *c, const pa_sink_info *i, int eol,
                               void *userdata) {
  struct sink_info_data *data = (struct sink_info_data *)userdata;
  if (eol > 0) {
    data->done = 1;
    pa_threaded_mainloop_signal(mloop, 0);
    return;
  }
  if (i) {
    if (i->volume.channels > 0) {
      pa_volume_t sum = 0;
      for (unsigned int j = 0; j < i->volume.channels; j++) {
        sum += i->volume.values[j];
      }
      double avg = (double)sum / i->volume.channels;
      data->volume = (avg / PA_VOLUME_NORM) * 100.0;
    }
    data->mute = i->mute;
  }
  data->done = 1;
  pa_threaded_mainloop_signal(mloop, 0);
}

int pulse_get_sink_info(const char *sink, double *volume, int *muted) {
  if (!context || !mloop)
    return -1;
  pa_threaded_mainloop_lock(mloop);
  struct sink_info_data data;
  memset(&data, 0, sizeof(data));
  data.done = 0;
  pa_operation *op = pa_context_get_sink_info_by_name(
      context, sink, sink_info_callback, &data);
  if (!op) {
    pa_threaded_mainloop_unlock(mloop);
    return -1;
  }
  while (!data.done)
    pa_threaded_mainloop_wait(mloop);
  pa_operation_unref(op);
  pa_threaded_mainloop_unlock(mloop);
  if (volume)
    *volume = data.volume;
  if (muted)
    *muted = data.mute;
  return 0;
}

int pulse_set_volume(const char *sink, double volume) {
  if (!context || !mloop)
    return -1;
  pa_threaded_mainloop_lock(mloop);
  pa_volume_t vol = (pa_volume_t)((volume / 100.0) * PA_VOLUME_NORM);
  pa_cvolume pcvolume;
  pa_cvolume_set(&pcvolume, 2, vol);
  int ret = 0;
  pa_operation *op =
      pa_context_set_sink_volume_by_name(context, sink, &pcvolume, NULL, NULL);
  if (!op) {
    ret = -1;
  } else {
    while (pa_operation_get_state(op) == PA_OPERATION_RUNNING)
      pa_threaded_mainloop_wait(mloop);
    pa_operation_unref(op);
  }
  pa_threaded_mainloop_unlock(mloop);
  return ret;
}

int pulse_set_mute(const char *sink, int mute) {
  if (!context || !mloop)
    return -1;
  pa_threaded_mainloop_lock(mloop);
  int ret = 0;
  pa_operation *op =
      pa_context_set_sink_mute_by_name(context, sink, mute, NULL, NULL);
  if (!op) {
    ret = -1;
  } else {
    while (pa_operation_get_state(op) == PA_OPERATION_RUNNING)
      pa_threaded_mainloop_wait(mloop);
    pa_operation_unref(op);
  }
  pa_threaded_mainloop_unlock(mloop);
  return ret;
}

static void subscription_callback(pa_context *c, pa_subscription_event_type_t t,
                                  uint32_t idx, void *userdata) {
  pa_subscription_event_type_t facility =
      t & PA_SUBSCRIPTION_EVENT_FACILITY_MASK;
  if (facility == PA_SUBSCRIPTION_EVENT_SINK ||
      facility == PA_SUBSCRIPTION_EVENT_SERVER) {
    sink_event_callback_cgo("", -1.0, -1);
  }
}

int pulse_subscribe(sink_event_callback_t callback) {
  if (!context || !mloop)
    return -1;
  pa_threaded_mainloop_lock(mloop);
  pa_context_set_subscribe_callback(context, subscription_callback, NULL);

  pa_operation *op = pa_context_subscribe(
      context, PA_SUBSCRIPTION_MASK_SINK | PA_SUBSCRIPTION_MASK_SERVER, NULL,
      NULL);
  if (!op) {
    pa_threaded_mainloop_unlock(mloop);
    return -1;
  }
  while (pa_operation_get_state(op) == PA_OPERATION_RUNNING)
    pa_threaded_mainloop_wait(mloop);
  pa_operation_unref(op);
  pa_threaded_mainloop_unlock(mloop);
  return 0;
}

void sink_event_callback_cgo(const char *sink, double volume, int mute) {
  goSinkEventCallback((char *)sink, volume, mute);
}
