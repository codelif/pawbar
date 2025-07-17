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
	"github.com/codelif/katnip"
	"github.com/codelif/pawbar/pkg/dbusmenukitty/menu"
	"github.com/codelif/gorsvg"
	"github.com/codelif/xdgicons"
	"github.com/codelif/xdgicons/missing"
	"github.com/fxamacker/cbor/v2"
	"golang.org/x/image/colornames"

	l "log"
)

var fgColor color.Color

func Leaf(k *katnip.Kitty, rw io.ReadWriter) int {
	dec := cbor.NewDecoder(rw)
	l.SetOutput(io.Discard)

	log.SetOutput(rw)
	log.SetLevel(log.LevelTrace)
	dataEvents := make(chan []menu.Item, 10)
	go func() {
		var layout []menu.Item
		for {
			if err := dec.Decode(&layout); err == nil {
				dataEvents <- layout
			}
		}
	}()

	vx, err := vaxis.New(vaxis.Options{EnableSGRPixels: true})
	if err != nil {
		return 1
	}

	// query foreground color and store (used for rendering icon SVGs later)
	c := vx.QueryForeground()
	rgb := c.Params()
	fgColor = color.RGBA{rgb[0], rgb[1], rgb[2], 255}

	win := vx.Window()
	w, h := win.Size()
	screenEvents := vx.Events()
	var j []menu.Item

	for {
		select {
		case ev := <-screenEvents:
			switch ev.(type) {
			case vaxis.Redraw:
				vx.Render()
			case vaxis.Resize:
				win = vx.Window()
				win.Clear()
				prevw, prevh := w, h
				w, h = win.Size()
				if prevw != w || prevh != h {
					log.Debug("dbusmenukitty: %d, %d\n", w, h)
					draw(win, j)
					vx.Render()
				}
				// fmt.Fprintf(rw, "wow")
				// fmt.Fprintln(rw, j)
			case vaxis.Mouse:
				l.Printf("%#v\n", ev.(vaxis.Mouse))
			}
		case ev := <-dataEvents:
			j = ev
			maxHorizontalLength := maxLengthLabel(j)
			maxVerticalLength := len(j)
			if w != maxHorizontalLength || h != maxVerticalLength {
				k.Resize(maxHorizontalLength+4, maxVerticalLength)
				continue
			}

			win.Clear()
			draw(win, j)
			vx.Render()
		}
	}
	return 1
}

func draw(win vaxis.Window, j []menu.Item) {
	arrowHeads := []rune{'◄', '►'}
	for i, v := range j {
		if v.Type == menu.ItemSeparator {
			w, _ := win.Size()
			win.Println(i, vaxis.Segment{
				Text:  strings.Repeat("─", w),
				Style: vaxis.Style{Attribute: vaxis.AttrDim},
			})
		} else {
			var style vaxis.Style
			var c color.Color
			c = fgColor
			prefix := "  "
			suffix := prefix

			if !v.Enabled {
				style.Attribute |= vaxis.AttrDim
				c = colornames.Gray
			}
			if v.HasChildren {
				prefix = string(arrowHeads[0]) + string(prefix[1])
			}

			win.Println(i, vaxis.Segment{Text: prefix + v.Label.Display + suffix, Style: style})
			if v.IconData != nil {
				// log.Println("creating image...")
				img, err := png.Decode(bytes.NewReader(v.IconData))
				if err != nil {
					log.Trace("png decode error:", err)
					img = missing.GenerateMissingIconBroken(32, fgColor)
				}

				kimg := win.Vx.NewKittyGraphic(img)
				// kimg.SetPadding(3)
				kimg.Resize(2, 1)
				iw, ih := kimg.CellSize()
				log.Trace("kitty image size: %d, %d", iw, ih)
				kimg.Draw(win.New(0, i, iw, ih))
			} else if v.IconName != "" {
				img, err := renderIcon(v.Icon, c)
				if err != nil {
					img = missing.GenerateMissingIcon(32, fgColor)
				}

				kimg := win.Vx.NewKittyGraphic(img)
				// kimg.SetPadding(5)
				kimg.Resize(2, 1)
				iw, ih := kimg.CellSize()
				log.Trace("kitty image size: %d, %d", iw, ih)
				kimg.Draw(win.New(0, i, iw, ih))
			}
		}
	}
}

func maxLengthLabel(labels []menu.Item) int {
	if len(labels) == 0 {
		return 0
	}

	maxLen := len(labels[0].Label.Display)
	for _, l := range labels[1:] {
		if curLen := len(l.Label.Display); curLen > maxLen {
			maxLen = curLen
		}
	}
	return maxLen
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
