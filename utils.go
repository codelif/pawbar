package main

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
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

func DrawHorizontal(scr tcell.Screen, x, y int, style tcell.Style, text string) {
	for _, char := range text {
		scr.SetContent(x, y, char, nil, style)
		x++
	}
}

func DrawVertical(scr tcell.Screen, x, y int, style tcell.Style, text string) {
	for _, char := range text {
		scr.SetContent(x, y, char, nil, style)
		y++
	}
}

func FillVertical(scr tcell.Screen, x, y int, style tcell.Style, r rune, n int) {
	DrawVertical(scr, x, y, style, strings.Repeat(string(r), n))
}

func FillHorizontal(scr tcell.Screen, x, y int, style tcell.Style, r rune, n int) {
	DrawHorizontal(scr, x, y, style, strings.Repeat(string(r), n))
}

func FillBox(scr tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, c rune) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}

	for y := y1; y <= y2; y++ {
		DrawHorizontal(scr, x1, y, style, strings.Repeat(string(c), x2-x1+1))
	}
}
