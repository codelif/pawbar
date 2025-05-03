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
	win.Clear()

	modev, l, r, err := config.InitModules(cfgPath)
	if err != nil {
		utils.Logger.Fatalln("Failed to init modules from config:", err)
	}

	screenEvents := vx.Events()

	w, h := win.Size()
	pw, ph := 0, 0
	mouseX, mouseY := 0, 0
	utils.Logger.Printf("Panel Size (cells): %d, %d\n", w, h)
	mouseShape := vaxis.MouseShapeDefault

	tui.Init(w, h, l, r)
	tui.FullRender(win)
	vx.Render()

	isRunning := true
	for isRunning {
		select {
		case ev := <-screenEvents:
			switch ev := ev.(type) {
			case vaxis.Resize:
				pw, ph = ev.XPixel, ev.YPixel
				win = vx.Window()
				w, h = win.Size()
				tui.Resize(w, h)
				tui.FullRender(win)
				vx.Render()
				utils.Logger.Printf("Panel Size: %d, %d\n", pw, ph)
			case vaxis.Redraw:
				tui.FullRender(win)
				vx.Render()
			case vaxis.Key:
				if ev.String() == "Ctrl+c" {
					isRunning = false
					vx.PostEvent(vaxis.QuitEvent{})
				}
			case vaxis.Mouse:
				mouseX, mouseY = ev.Col, ev.Row

				if mouseY != 0 {
					continue
				}
				c := tui.State()[mouseX]
				if c.Mod != nil {
					_, send := c.Mod.Channels()
					send <- modules.Event{Cell: c, VaxisEvent: ev}
				}
				updateMouseShape(vx, c, &mouseShape, true)
			case vaxis.QuitEvent:
				utils.Logger.Printf("Received exit signal\n")
				isRunning = false
			}
		case m := <-modev:
			utils.Logger.Println("render:", m.Name())
			tui.PartialRender(win, m)
			vx.Render()
		}
	}
	return 0
}

func updateMouseShape(
	vx *vaxis.Vaxis,
	ec modules.EventCell,
	old *vaxis.MouseShape,
	render bool,
) {
	target := ec.MouseShape
	if target == "" {
		target = vaxis.MouseShapeDefault
	}
	if ec.Mod == nil {
		target = vaxis.MouseShapeDefault
	}

	if *old == target {
		return
	}

	*old = target
	utils.Logger.Printf("mouse shape: %s\n", *old)
	vx.SetMouseShape(target)

	if render {
		vx.Render()
	}
}
