#include <stdint.h>

int pulse_init();
int pulse_get_default_sink(char *sink, int sink_len);
int pulse_get_sink_info(const char *sink, double *volume, int *muted);
int pulse_set_volume(const char *sink, double volume);
// (mute=1, unmute=0)
int pulse_set_mute(const char *sink, int mute);

typedef void (*sink_event_callback_t)(const char *sink, double volume,
                                      int mute);

int pulse_subscribe(sink_event_callback_t callback);
void sink_event_callback_cgo(const char *sink, double volume, int mute);
extern void goSinkEventCallback(char *sink, double volume, int mute);
