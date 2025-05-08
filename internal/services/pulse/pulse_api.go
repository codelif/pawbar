package pulse

/*
#cgo pkg-config: libpulse
#include "pulse_wrapper.h"
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"unsafe"
)

type SinkInfo struct {
	Volume float64
	Muted  bool
}

type SinkEvent struct {
	Sink   string
	Volume float64
	Muted  bool
}

func init_pulse() error {
	if C.pulse_init() != 0 {
		return errors.New("failed to initialize PulseAudio")
	}
	return nil
}

func getDefaultSink() (string, error) {
	buf := make([]C.char, 256)
	if C.pulse_get_default_sink(&buf[0], C.int(len(buf))) != 0 {
		return "", errors.New("failed to get default sink")
	}
	return C.GoString(&buf[0]), nil
}

func getSinkInfo(sink string) (SinkInfo, error) {
	csink := C.CString(sink)
	defer C.free(unsafe.Pointer(csink))
	var volume C.double
	var muted C.int
	if C.pulse_get_sink_info(csink, &volume, &muted) != 0 {
		return SinkInfo{}, errors.New("failed to get sink info")
	}
	return SinkInfo{Volume: float64(volume), Muted: muted != 0}, nil
}

func setVolume(sink string, volume float64) error {
	csink := C.CString(sink)
	defer C.free(unsafe.Pointer(csink))
	if C.pulse_set_volume(csink, C.double(volume)) != 0 {
		return errors.New("failed to set volume")
	}
	return nil
}

func setMute(sink string, mute bool) error {
	csink := C.CString(sink)
	defer C.free(unsafe.Pointer(csink))
	var imute C.int
	if mute {
		imute = 1
	} else {
		imute = 0
	}
	if C.pulse_set_mute(csink, imute) != 0 {
		return errors.New("failed to set mute")
	}
	return nil
}

var sinkEventChan chan SinkEvent

//export goSinkEventCallback
func goSinkEventCallback(cSink *C.char, volume C.double, muted C.int) {
	sinkStr := C.GoString(cSink)
	if sinkStr == "" || float64(volume) < 0 {
		go func() {
			s, err := getDefaultSink()
			if err != nil {
				return
			}
			info, err := getSinkInfo(s)
			if err != nil {
				return
			}
			if sinkEventChan != nil {
				sinkEventChan <- SinkEvent{Sink: s, Volume: info.Volume, Muted: info.Muted}
			}
		}()
	} else {
		if sinkEventChan != nil {
			sinkEventChan <- SinkEvent{Sink: sinkStr, Volume: float64(volume), Muted: muted != 0}
		}
	}
}

func monitor() (<-chan SinkEvent, error) {
	sinkEventChan = make(chan SinkEvent, 10)

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		// drive the callbacks until pulse service will unsubscribe
		if C.pulse_subscribe((C.sink_event_callback_t)(C.sink_event_callback_cgo)) != 0 {
			fmt.Fprintln(os.Stderr, "pulse_subscribe failed")
		}
	}()

	return sinkEventChan, nil
}
