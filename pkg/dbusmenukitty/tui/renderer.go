package tui

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	l "log"
	"os"
	"path/filepath"
	"strings"

	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/log"
	"github.com/codelif/gorsvg"
	"github.com/nekorg/pawbar/pkg/dbusmenukitty/menu"
	"github.com/codelif/xdgicons"
	"github.com/codelif/xdgicons/missing"
	"golang.org/x/image/colornames"
)

type Renderer struct {
	win vaxis.Window
}

func NewRenderer(win vaxis.Window) *Renderer {
	return &Renderer{win: win}
}

func (r *Renderer) drawBackground(row int, style vaxis.Style) {
	w, _ := r.win.Size()
	r.win.Println(row, vaxis.Segment{
		Text:  strings.Repeat(" ", w),
		Style: style,
	})
}

func (r *Renderer) getItemStyle(item *menu.Item, row int, state *MenuState) vaxis.Style {
	var style vaxis.Style

	if row == state.mouseY && state.mouseOnSurface {
		if state.mousePressed {
			style.Background = vaxis.ColorBlue
		} else {
			style.Background = vaxis.ColorGray
		}
	}

	if !item.Enabled {
		style.Background = 0
		style.Attribute |= vaxis.AttrDim
	}

	return style
}

func (r *Renderer) renderIcon(item *menu.Item, row int, defaultColor color.Color) (prefixAdd string) {
	var img image.Image
	var err error

	if item.IconData != nil {
		img, err = png.Decode(bytes.NewReader(item.IconData))
		if err != nil {
			log.Trace("png decode error:", err)
			img = missing.GenerateMissingIconBroken(iconSize, defaultColor)
		}
	} else if item.IconName != "" {
		img, err = renderIcon(item.Icon, defaultColor)
		if err != nil {
			img = missing.GenerateMissingIcon(iconSize, defaultColor)
		}
	} else {
		return ""
	}

	kimg := r.win.Vx.NewKittyGraphic(img)
	kimg.Resize(iconCellWidth, iconCellHeight)
	iw, ih := kimg.CellSize()
	log.Trace("kitty image size: %d, %d", iw, ih)
	kimg.Draw(r.win.New(menuPadding, row, iw, ih))

	return strings.Repeat(" ", iconSpacing)
}

func (r *Renderer) drawItem(item *menu.Item, row int, state *MenuState, showIcons bool) {
	style := r.getItemStyle(item, row, state)
	defaultColor := fgColor

	if !item.Enabled {
		defaultColor = colornames.Gray
	}

	// Draw background
	r.drawBackground(row, style)

	// Build prefix
	prefix := strings.Repeat(" ", menuPadding)
	if item.HasChildren {
		prefix = string(arrowHeads[0]) + prefix[1:]
	}

	// Add icon if present and rendering full menu
	if showIcons {
		prefix += r.renderIcon(item, row, defaultColor)
	} else if item.IconData != nil || item.IconName != "" {
		// Reserve space for icon in fast draw
		prefix += strings.Repeat(" ", iconSpacing)
	}

	// Draw text
	suffix := strings.Repeat(" ", menuPadding)
	text := prefix + item.Label.Display + suffix
	r.win.Println(row, vaxis.Segment{Text: text, Style: style})
}

func (r *Renderer) drawSeparator(row int) {
	w, _ := r.win.Size()
	r.win.Println(row, vaxis.Segment{
		Text:  strings.Repeat("─", w),
		Style: vaxis.Style{Attribute: vaxis.AttrDim},
	})
}

func (r *Renderer) drawMenu(items []menu.Item, state *MenuState, showIcons bool) {
	for i, item := range items {
		if item.Type == menu.ItemSeparator {
			r.drawSeparator(i)
		} else {
			r.drawItem(&item, i, state, showIcons)
		}
	}
}

func renderIcon(icon xdgicons.Icon, c color.Color) (image.Image, error) {
	l.Printf("%v\n", icon)

	if icon.Path == "" {
		return nil, fmt.Errorf("no file path provided")
	}

	isSymbolic := strings.HasSuffix(icon.Name, "-symbolic")
	ext := strings.ToLower(filepath.Ext(icon.Path))

	f, err := os.Open(icon.Path)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", icon.Path, err)
	}
	defer f.Close()

	switch ext {
	case ".svg":
		if isSymbolic {
			return gorsvg.DecodeWithColor(f, 48, 48, c)
		}
		return gorsvg.Decode(f, 48, 48)

	case ".png":
		return png.Decode(f)

	default:
		return nil, fmt.Errorf("unsupported image format: %s", ext)
	}
}
