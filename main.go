package main

import (
	"io"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gdamore/tcell/v2"
)

var logger *log.Logger

func main() {
	var Fd *os.File
	logger = log.New(io.Discard, "", 0)
	if len(os.Args) > 1 {
		logger = log.New(&notifyio, "", 0)
		fi, err := os.Lstat(os.Args[1])
		if err != nil {
			logger.Fatalln("There was an error accessing the char device path.")
		}
		if fi.Mode()&fs.ModeCharDevice == 0 {
			logger.Fatalln("The given path is not a char device.")
		}

		device := os.Args[1]
		Fd, err = os.OpenFile(device, os.O_WRONLY, 0620)
		if err != nil {
			logger.Fatalln("There was an error opening the char device.")
		}
		logger = log.New(Fd, "", log.LstdFlags)
		defer Fd.Close()
	}
	exit_code := MainLoop()
	Fd.Close()

	os.Exit(exit_code)
}

func MainLoop() int {
	scr, err := tcell.NewScreen()
	if err != nil {
		logger.Println("There was an error creating a Screen.")
		return 1
	}

	err = scr.Init()
	if err != nil {
		logger.Println("There was an error initializing the Screen.")
		return 1
	}

	defer func() {
		if_panic := recover()
		scr.Fini()
		if if_panic != nil {
			panic(if_panic)
		}
	}()

	scr.EnableMouse()

	style := tcell.StyleDefault
	w, h := scr.Size()
	logger.Println("Panel Size:", strconv.Itoa(w)+", "+strconv.Itoa(h))
	scr.SetStyle(style)
	scr.Clear()

	exit_signal := make(chan os.Signal, 1)
	signal.Notify(exit_signal, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	modev, l, r := StartModules()

	screen_events := make(chan tcell.Event)
	quit := make(chan struct{})
	go scr.ChannelEvents(screen_events, quit)

	cells := make([]EventCell, w)
	Refresh(scr, l, r, cells)
	scr.Show()
	// m := MakeIMap(w)
	running := true
	for running {
		select {
		case ev := <-screen_events:
			switch ev := ev.(type) {
			case *tcell.EventResize:
				w, h = scr.Size()
			case *tcell.EventKey:
				logger.Printf("Key: %s\n", ev.Name())
				if ev.Key() == tcell.KeyCtrlC {
					exit_signal <- os.Interrupt
				}
			case *tcell.EventMouse:
				x, y := ev.Position()

				if y != 0 {
					continue
				}
				c := cells[x]
				if c.m != nil {
					_, send := c.m.Channels()
					send <- Event{c, ev}
				}

			case *tcell.EventPaste:
				logger.Printf("Paste: %t, %t\n", ev.Start(), ev.End())
			}
		case <-modev:
			Refresh(scr, l, r, cells)
			scr.Show()
		case s := <-exit_signal:
			logger.Printf("Received exit signal: %s\n", s.String())
			quit <- struct{}{}
			running = false
		}

	}

	return 0
}

func Refresh(scr tcell.Screen, l, r []Module, cells []EventCell) {
	w, _ := scr.Size()

	for i := range w {
		cells[i] = EventCell{' ', DEFAULT, "", nil}
		scr.SetContent(i, 0, ' ', nil, DEFAULT)
	}

	p := 0
	for _, mod := range l {
		for _, c := range mod.Render() {
			// logger.Printf("%s: [%d]: '%c', '%s'\n", mod.Name(), p, c.c, c.m.Name())
			scr.SetContent(p, 0, c.c, nil, c.style)
			cells[p] = c
			p++
		}
	}

	p = 0
	for _, mod := range r {
		mod_render := mod.Render()
		len_mod := len(mod_render)
		for i := range len_mod {
			c := mod_render[len_mod-i-1]
			scr.SetContent(w-p-1, 0, c.c, nil, c.style)
			cells[w-p-1] = c
			p++
		}
	}

}
