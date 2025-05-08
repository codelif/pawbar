package utils

import (
	"cmp"
	"errors"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	NotifyIO NotificationWriter
	Logger   *log.Logger
)

type NotificationWriter struct{}

func (w *NotificationWriter) Write(p []byte) (n int, err error) {
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

func Clamp[T cmp.Ordered](n, low, high T) T {
	if n < low {
		return low
	}

	if n > high {
		return high
	}

	return n
}

func InitLogger() (*log.Logger, *os.File) {
	var Fd *os.File
	Logger = log.New(io.Discard, "", 0)
	if len(os.Args) > 1 {
		Logger = log.New(&NotifyIO, "", 0)
		fi, err := os.Lstat(os.Args[1])
		if err != nil {
			Logger.Fatalln("There was an error accessing the char device path.")
		}
		if fi.Mode()&fs.ModeCharDevice == 0 {
			Logger.Fatalln("The given path is not a char device.")
		}

		device := os.Args[1]
		Fd, err = os.OpenFile(device, os.O_WRONLY, 0o620)
		if err != nil {
			Logger.Fatalln("There was an error opening the char device.")
		}
		Logger = log.New(Fd, "", log.LstdFlags)
	}
	return Logger, Fd
}

func set_font_size(size int) {
	exec.Command("kitty", "@", "set-font-size", strconv.Itoa(size)).Run()
}
