package main

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/tui"
	"github.com/codelif/pawbar/internal/utils"
	"github.com/gdamore/tcell/v2"
)

func main() {
	_, Fd := utils.InitLogger()
	defer Fd.Close()

  configFile := os.Getenv("HOME") + "/.config/pawbar/pawbar.yaml"
	exitCode := mainLoop(configFile)

	os.Exit(exitCode)
}

func mainLoop(cfgPath string) int {
	scr, err := tcell.NewScreen()
	if err != nil {
		utils.Logger.Println("There was an error creating a Screen.")
		return 1
	}

	err = scr.Init()
	if err != nil {
		utils.Logger.Println("There was an error initializing the Screen.")
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
	utils.Logger.Println("Panel Size:", strconv.Itoa(w)+", "+strconv.Itoa(h))
	scr.SetStyle(style)
	scr.Clear()

	exit_signal := make(chan os.Signal, 1)
	signal.Notify(exit_signal, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	modev, l, r, err := config.InitModules(cfgPath)
  if err != nil {
    utils.Logger.Fatalln("Failed to init modules from config:", err)
  }

	screenEvents := make(chan tcell.Event)
	quitEventsChan := make(chan struct{})
	go scr.ChannelEvents(screenEvents, quitEventsChan)

	renderCells := make([]modules.EventCell, w)

	isRunning := true
	for isRunning {
		select {
		case ev := <-screenEvents:
			switch ev := ev.(type) {
			case *tcell.EventResize:
				w, h = scr.Size()
				utils.Logger.Println("Panel Size:", strconv.Itoa(w)+", "+strconv.Itoa(h))
				renderCells = make([]modules.EventCell, w)
				tui.RenderBar(scr, l, r, renderCells)
				scr.Show()
			case *tcell.EventKey:
				utils.Logger.Printf("Key: %s\n", ev.Name())
				if ev.Key() == tcell.KeyCtrlC {
					exit_signal <- os.Interrupt
				}
			case *tcell.EventMouse:
				x, y := ev.Position()

				if y != 0 {
					continue
				}
				c := renderCells[x]
				if c.Mod != nil {
					_, send := c.Mod.Channels()
					send <- modules.Event{Cell: c, TcellEvent: ev}
				}

			case *tcell.EventPaste:
				utils.Logger.Printf("Paste: %t, %t\n", ev.Start(), ev.End())
			}
    case m:=<-modev:
      utils.Logger.Println("Received render event from:", m.Name())
			tui.RenderBar(scr, l, r, renderCells)
			scr.Show()
		case s := <-exit_signal:
			utils.Logger.Printf("Received exit signal: %s\n", s.String())
			quitEventsChan <- struct{}{}
			isRunning = false
		}

	}

	return 0
}
