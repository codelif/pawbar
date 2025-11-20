package tui

import (
	"image/color"
	"io"
	"time"

	l "log"

	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/log"
	"github.com/nekorg/katnip"
	"github.com/nekorg/pawbar/pkg/dbusmenukitty/menu"
	"github.com/fxamacker/cbor/v2"
)

const (
	hoverActivationTimeout = 200 * time.Millisecond
	iconSize               = 32
	iconCellWidth          = 2
	iconCellHeight         = 1
	menuPadding            = 2
	iconSpacing            = 2
)

var (
	fgColor    color.Color
	arrowHeads = []rune{'◄', '►'}
)

func Leaf(k *katnip.Kitty, rw io.ReadWriter) int {
	// device := "/dev/pts/8"
	// Fd, err := os.OpenFile(device, os.O_WRONLY, 0o620)
	// if err == nil {
	// 	log.SetOutput(Fd)
	// 	log.SetLevel(log.LevelTrace)
	// }
	dec := cbor.NewDecoder(rw)
	enc := cbor.NewEncoder(rw)

	// Disable logging
	l.SetOutput(io.Discard)
	// log.SetOutput(io.Discard)
	// log.SetLevel(log.LevelInfo)

	// Setup message queue
	msgQueue := make(chan menu.Message, 10)
	go func() {
		var msg menu.Message
		for {
			if err := dec.Decode(&msg); err == nil {
				msgQueue <- msg
			}
		}
	}()

	// Initialize vaxis
	vx, err := vaxis.New(vaxis.Options{EnableSGRPixels: true})
	if err != nil {
		return 1
	}

	// Query and store foreground color for icon rendering
	c := vx.QueryForeground()
	rgb := c.Params()
	fgColor = color.RGBA{rgb[0], rgb[1], rgb[2], 255}

	// Initialize state and handlers
	state := NewMenuState()
	messageHandler := NewMessageHandler(enc, state)

	win := vx.Window()
	renderer := NewRenderer(win)

	{
		winSize := vx.Size()
		state.size = menu.Size{
			Cols:    winSize.Cols,
			Rows:    winSize.Rows,
			XPixels: winSize.XPixel,
			YPixels: winSize.YPixel,
		}
		state.ppc = menu.PPC{
			X: float64(winSize.XPixel) / float64(winSize.Cols),
			Y: float64(winSize.YPixel) / float64(winSize.Rows),
		}
	}

	k.Show()

	for {
		select {
		case ev := <-vx.Events():
			switch ev := ev.(type) {
			case vaxis.Redraw:
				vx.Render()

			case vaxis.Resize:
				win = vx.Window()
				renderer = NewRenderer(win)
				win.Clear()
				state.size = menu.Size{
					Cols:    ev.Cols,
					Rows:    ev.Rows,
					XPixels: ev.XPixel,
					YPixels: ev.YPixel,
				}

				state.ppc = menu.PPC{
					X: float64(ev.XPixel) / float64(ev.Cols),
					Y: float64(ev.YPixel) / float64(ev.Rows),
				}

				log.Debug("dbusmenukitty: %d, %d\n", state.size.Cols, state.size.Rows)
				renderer.drawMenu(state.items, state, true)
				vx.Render()

			case vaxis.Mouse:
				l.Printf("%#v\n", ev)
				state.mousePixelX, state.mousePixelY = ev.XPixel, ev.YPixel
				switch ev.EventType {
				case vaxis.EventLeave:
					state.mouseOnSurface = false

				case vaxis.EventMotion:
					messageHandler.handleMouseMotion(ev.Col, ev.Row)

				case vaxis.EventPress:
					if ev.Button == vaxis.MouseLeftButton {
						state.mousePressed = true
					}

				case vaxis.EventRelease:
					if ev.Button == vaxis.MouseLeftButton {
						state.mousePressed = false
						if state.isSelectableItem(state.mouseY) {
							messageHandler.handleItemClick(&state.items[state.mouseY])
						}
					}
				}

				renderer.drawMenu(state.items, state, false) // Fast draw (text only)
				vx.Render()

			case vaxis.Key:
				if ev.EventType == vaxis.EventPress {
					switch ev.Keycode {
					case vaxis.KeyEsc, vaxis.KeyLeft:
						return 0

					case vaxis.KeyEnter:
						if state.isSelectableItem(state.mouseY) {
							messageHandler.handleItemClick(&state.items[state.mouseY])
						}

					case vaxis.KeyUp:
						state.navigateUp()
						messageHandler.handleKeyNavigation(true)
						renderer.drawMenu(state.items, state, false)
						vx.Render()

					case vaxis.KeyDown:
						state.navigateDown()
						messageHandler.handleKeyNavigation(true)
						renderer.drawMenu(state.items, state, false)
						vx.Render()

					case vaxis.KeyRight:
						if item := state.getCurrentItem(); item != nil && item.HasChildren && item.Enabled {
							messageHandler.sendMessage(menu.MsgSubmenuRequested, item.Id,
								state.mousePixelX, state.mousePixelY)
						}
					}
				}
			}

		case msg := <-msgQueue:
			switch msg.Type {
			case menu.MsgMenuClose:
				return 0

			case menu.MsgMenuUpdate:
				state.items = msg.Payload.Menu
				maxHorizontalLength := menu.MaxLengthLabel(state.items) + 4
				maxVerticalLength := len(state.items)

				log.Info("leaf: %d %d actual: %d %d\n", maxHorizontalLength, maxVerticalLength, state.size.Cols, state.size.Rows)

				win.Clear()
				renderer.drawMenu(state.items, state, true)
				vx.Render()

				if state.size.Cols != maxHorizontalLength || state.size.Rows != maxVerticalLength {
					log.Info("resizing window to fit menu")
					k.Resize(maxHorizontalLength, maxVerticalLength)
					continue
				}
			}
		}
	}
}
