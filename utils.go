package main

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type NotifWriter struct{}

func (w *NotifWriter) Write(p []byte) (n int, err error) {
	cmd := exec.Command("notify-send", strings.TrimRight(string(p), "\n"))

	done := make(chan error)
	go func() {
		err := cmd.Run()
		done <- err
	}()

	select {
	case <-time.After(1 * time.Second):
		return 0, errors.New("Command timed out.")
	case d := <-done:
		if d != nil {
			return 0, d
		}
		return len(p), d
	}

	return 0, errors.New("utils:NotifWriter:Write: Unreachable")
}

var notifyio NotifWriter

func set_font_size(size int) {
	exec.Command("kitty", "@", "set-font-size", strconv.Itoa(size)).Run()
}
