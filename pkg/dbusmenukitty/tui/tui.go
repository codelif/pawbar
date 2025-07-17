package tui

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/log"
	"github.com/codelif/gorsvg"
	"github.com/codelif/katnip"
	"github.com/codelif/pawbar/pkg/dbusmenukitty/menu"
	"github.com/codelif/xdgicons"
	"github.com/codelif/xdgicons/missing"
	"github.com/fxamacker/cbor/v2"
	"golang.org/x/image/colornames"

	l "log"
)

var fgColor color.Color

func Leaf(k *katnip.Kitty, rw io.ReadWriter) int {
	dec := cbor.NewDecoder(rw)
	enc := cbor.NewEncoder(rw)
	l.SetOutput(io.Discard)

	log.SetOutput(io.Discard)
	log.SetLevel(log.LevelInfo)
	msgQueue := make(chan menu.Message, 10)
	go func() {
		var msg menu.Message
		for {
			if err := dec.Decode(&msg); err == nil {
				msgQueue <- msg
			}
		}
	}()

	vx, err := vaxis.New(vaxis.Options{EnableSGRPixels: true})
	if err != nil {
		return 1
	}

	defer func() {
		err := recover()
		log.Info("ERROR: %v\n", err)
		vx.Close()
	}()
	// query foreground color and store (used for rendering icon SVGs later)
	c := vx.QueryForeground()
	rgb := c.Params()
	fgColor = color.RGBA{rgb[0], rgb[1], rgb[2], 255}

	win := vx.Window()
	w, h := win.Size()
	mouseY := -1
	mousePixelX, mousePixelY := -1, -1
	lastMouseY := -1
	mousePressed := false
	screenEvents := vx.Events()
	var menuItems []menu.Item

	for {
		select {
		case ev := <-screenEvents:
			switch ev := ev.(type) {
			case vaxis.Redraw:
				vx.Render()
			case vaxis.Resize:
				win = vx.Window()
				win.Clear()
				w, h = win.Size()
				log.Debug("dbusmenukitty: %d, %d\n", w, h)
				draw(win, menuItems, mouseY, mousePressed)
				vx.Render()
			case vaxis.Mouse:
				l.Printf("%#v\n", ev)

				switch ev.EventType {
				case vaxis.EventLeave:
					if lastMouseY != -1 {
						// Send unhover event for last item
						msg := menu.Message{
							Type: menu.MsgItemHovered,
							Payload: menu.MessagePayload{
								ItemId: 0, // 0 means no item hovered
							},
						}
						enc.Encode(msg)
						lastMouseY = -1
					}
					mouseY = -1
				case vaxis.EventMotion:
					mousePixelX, mousePixelY = ev.XPixel, ev.YPixel
					if mouseY == ev.Row {
						continue
					}
					mouseY = ev.Row

					// Send hover event if row changed and is valid
					if mouseY >= 0 && mouseY < len(menuItems) && menuItems[mouseY].Type != menu.ItemSeparator {
						if lastMouseY != mouseY {
							msg := menu.Message{
								Type: menu.MsgItemHovered,
								Payload: menu.MessagePayload{
									ItemId: menuItems[mouseY].Id,
								},
							}
							enc.Encode(msg)
							lastMouseY = mouseY
						}
					}
				case vaxis.EventPress:
					if ev.Button != vaxis.MouseLeftButton {
						continue
					}
					mousePressed = true
				case vaxis.EventRelease:
					if ev.Button != vaxis.MouseLeftButton {
						continue
					}
					mousePressed = false

					// Send click event if on valid item
					if mouseY >= 0 && mouseY < len(menuItems) && menuItems[mouseY].Type != menu.ItemSeparator && menuItems[mouseY].Enabled {
						item := menuItems[mouseY]

						if item.HasChildren {
							// Send submenu request
							msg := menu.Message{
								Type: menu.MsgSubmenuRequested,
								Payload: menu.MessagePayload{
									ItemId: item.Id,
									PixelX: mousePixelX,
									PixelY: mousePixelY,
								},
							}
							enc.Encode(msg)
						} else {
							// Send click event
							msg := menu.Message{
								Type: menu.MsgItemClicked,
								Payload: menu.MessagePayload{
									ItemId: item.Id,
								},
							}
							enc.Encode(msg)
						}
					}
				default:
					continue
				}

				drawFast(win, menuItems, mouseY, mousePressed) // renders only text
				vx.Render()
			case vaxis.Key:
				switch ev.EventType {
				case vaxis.EventPress:
					switch ev.Keycode {
					case vaxis.KeyEsc:
						return 0
					case vaxis.KeyLeft:
						return 0
					case vaxis.KeyEnter:
						if mouseY >= 0 && mouseY < len(menuItems) && menuItems[mouseY].Type != menu.ItemSeparator && menuItems[mouseY].Enabled {
							item := menuItems[mouseY]

							if item.HasChildren {
								msg := menu.Message{
									Type: menu.MsgSubmenuRequested,
									Payload: menu.MessagePayload{
										ItemId: item.Id,
										PixelX: mousePixelX,
										PixelY: mousePixelY,
									},
								}
								enc.Encode(msg)
							} else {
								msg := menu.Message{
									Type: menu.MsgItemClicked,
									Payload: menu.MessagePayload{
										ItemId: item.Id,
									},
								}
								enc.Encode(msg)
							}
						}
					case vaxis.KeyUp:
						if mouseY > 0 {
							mouseY--
							for mouseY > 0 && menuItems[mouseY].Type == menu.ItemSeparator {
								mouseY--
							}
							drawFast(win, menuItems, mouseY, mousePressed)
							vx.Render()

							if mouseY >= 0 && mouseY < len(menuItems) {
								msg := menu.Message{
									Type: menu.MsgItemHovered,
									Payload: menu.MessagePayload{
										ItemId: menuItems[mouseY].Id,
									},
								}
								enc.Encode(msg)
								lastMouseY = mouseY
							}
						}
					case vaxis.KeyDown:
						if mouseY < len(menuItems)-1 {
							mouseY++
							for mouseY < len(menuItems)-1 && menuItems[mouseY].Type == menu.ItemSeparator {
								mouseY++
							}
							drawFast(win, menuItems, mouseY, mousePressed)
							vx.Render()

							if mouseY >= 0 && mouseY < len(menuItems) {
								msg := menu.Message{
									Type: menu.MsgItemHovered,
									Payload: menu.MessagePayload{
										ItemId: menuItems[mouseY].Id,
									},
								}
								enc.Encode(msg)
								lastMouseY = mouseY
							}
						}
					}
				}
			}
		case msg := <-msgQueue:
			switch msg.Type {
			case menu.MsgMenuUpdate:
				menuItems = msg.Payload.Menu
				maxHorizontalLength := menu.MaxLengthLabel(menuItems) + 4
				maxVerticalLength := len(menuItems)
				log.Info("leaf: %d %d actual: %d %d\n", maxHorizontalLength, maxVerticalLength, w, h)

				win.Clear()
				draw(win, menuItems, mouseY, mousePressed)
				vx.Render()

				if w != maxHorizontalLength || h != maxVerticalLength {
					k.Resize(maxHorizontalLength, maxVerticalLength)
					continue
				}
			}
		}
	}
	return 1
}

func drawFast(win vaxis.Window, items []menu.Item, mouseY int, mousePressed bool) {
	arrowHeads := []rune{'◄', '►'}
	w, _ := win.Size()

	for i, v := range items {
		if v.Type != menu.ItemSeparator {
			var style vaxis.Style
			prefix := "  "
			suffix := prefix

			if i == mouseY {
				if mousePressed {
					style.Background = vaxis.ColorBlue
				} else {
					style.Background = vaxis.ColorGray
				}
			}

			if !v.Enabled {
				style.Background = 0
				style.Attribute |= vaxis.AttrDim
			}
			if v.HasChildren {
				prefix = string(arrowHeads[0]) + string(prefix[1])
			}

			if v.IconData != nil {
				prefix += "  "
			} else if v.IconName != "" {
				prefix += "  "
			}

			win.Println(i, vaxis.Segment{Text: strings.Repeat(" ", w), Style: style})
			win.Println(i, vaxis.Segment{Text: prefix + v.Label.Display + suffix, Style: style})
		}
	}
}

func draw(win vaxis.Window, items []menu.Item, mouseY int, mousePressed bool) {
	w, _ := win.Size()
	arrowHeads := []rune{'◄', '►'}
	for i, v := range items {
		if v.Type == menu.ItemSeparator {
			w, _ := win.Size()
			win.Println(i, vaxis.Segment{
				Text:  strings.Repeat("─", w),
				Style: vaxis.Style{Attribute: vaxis.AttrDim},
			})
		} else {
			var style vaxis.Style
			defaultColor := fgColor
			prefix := "  "
			suffix := prefix

			if i == mouseY {
				if mousePressed {
					style.Background = vaxis.ColorBlue
				} else {
					style.Background = vaxis.ColorGray
				}
			}

			if !v.Enabled {
				style.Background = 0
				style.Attribute |= vaxis.AttrDim
				defaultColor = colornames.Gray
			}
			if v.HasChildren {
				prefix = string(arrowHeads[0]) + string(prefix[1])
			}

			if v.IconData != nil {
				img, err := png.Decode(bytes.NewReader(v.IconData))
				if err != nil {
					log.Trace("png decode error:", err)
					img = missing.GenerateMissingIconBroken(32, defaultColor)
				}

				kimg := win.Vx.NewKittyGraphic(img)
				kimg.Resize(2, 1)
				iw, ih := kimg.CellSize()
				log.Trace("kitty image size: %d, %d", iw, ih)
				kimg.Draw(win.New(2, i, iw, ih))
				prefix += "  "
			} else if v.IconName != "" {
				img, err := renderIcon(v.Icon, defaultColor)
				if err != nil {
					img = missing.GenerateMissingIcon(32, defaultColor)
				}

				kimg := win.Vx.NewKittyGraphic(img)
				kimg.Resize(2, 1)
				iw, ih := kimg.CellSize()
				log.Trace("kitty image size: %d, %d", iw, ih)
				kimg.Draw(win.New(2, i, iw, ih))
				prefix += "  "
			}
			win.Println(i, vaxis.Segment{Text: strings.Repeat(" ", w), Style: style})
			win.Println(i, vaxis.Segment{Text: prefix + v.Label.Display + suffix, Style: style})
		}
	}
}

func renderIcon(icon xdgicons.Icon, c color.Color) (img image.Image, err error) {
	l.Printf("%v\n", icon)
	if icon.Path == "" {
		return nil, fmt.Errorf("no file path")
	}
	isSymbolic := strings.HasSuffix(icon.Name, "-symbolic")

	ext := strings.ToLower(filepath.Ext(icon.Path))
	switch ext {
	case ".svg":
		px := 48
		f, err := os.Open(icon.Path)
		if err != nil {
			return nil, fmt.Errorf("error opening png file: %w", err)
		}

		if isSymbolic {
			img, err = gorsvg.DecodeWithColor(f, px, px, c)
		} else {
			img, err = gorsvg.Decode(f, px, px)
		}
		if err != nil {
			return nil, fmt.Errorf("error rendering svg: %w", err)
		}
	case ".png":
		f, err := os.Open(icon.Path)
		if err != nil {
			return nil, fmt.Errorf("error opening png file: %w", err)
		}
		img, err = png.Decode(f)
		if err != nil {
			return nil, fmt.Errorf("error decoding png: %w", err)
		}
	default:
		return nil, fmt.Errorf("no supported image files")
	}

	return img, nil
}
