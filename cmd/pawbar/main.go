package main

import (
	"os"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
	"github.com/codelif/pawbar/internal/tui"
	"github.com/codelif/pawbar/internal/utils"
)

func main() {
	_, Fd := utils.InitLogger()
	defer Fd.Close()

	configFile := os.Getenv("HOME") + "/.config/pawbar/pawbar.yaml"
	exitCode := mainLoop(configFile)

	os.Exit(exitCode)
}

func mainLoop(cfgPath string) int {
	vx, err := vaxis.New(vaxis.Options{EnableSGRPixels: true})
	if err != nil {
		utils.Logger.Println("There was an error initializing Vaxis.")
		return 1
	}
	defer vx.Close() // no need for recover since its done in vaxis

	win := vx.Window()
	w, h := win.Size()
	pw, ph := 0, 0
	utils.Logger.Printf("Panel Size (cells): %d, %d\n", w, h)
	win.Clear()

	modev, l, r, err := config.InitModules(cfgPath)
	if err != nil {
		utils.Logger.Fatalln("Failed to init modules from config:", err)
	}

	screenEvents := vx.Events()

	renderCells := make([]modules.EventCell, w)
	tui.RenderBar(win, l, r, nil, renderCells)

	isRunning := true
	for isRunning {
		select {
		case ev := <-screenEvents:
			switch ev := ev.(type) {
			case vaxis.Resize:
				w, h = ev.Cols, ev.Rows
				pw, ph = ev.XPixel, ev.YPixel
				utils.Logger.Printf("Panel Size: %d, %d\n", pw, ph)
			case vaxis.Redraw:
				renderCells = make([]modules.EventCell, w)
				tui.RenderBar(win, l, r, nil, renderCells)
				vx.Render()
			case vaxis.Key:
				if ev.String() == "Ctrl+c" {
					isRunning = false
					vx.PostEvent(vaxis.QuitEvent{})
				}
			case vaxis.Mouse:
				x, y := ev.Col, ev.Row

				if y != 0 {
					continue
				}
				c := renderCells[x]
				if c.Mod != nil {
					_, send := c.Mod.Channels()
					send <- modules.Event{Cell: c, VaxisEvent: ev}
				}
			case vaxis.QuitEvent:
				utils.Logger.Printf("Received exit signal\n")
				isRunning = false
			}
		case m := <-modev:
			utils.Logger.Println("Received render event from:", m.Name())
			tui.RenderBar(win, l, r, m, renderCells)
			vx.Render()
		}

	}
	return 0
}
