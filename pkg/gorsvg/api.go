package gorsvg

// #cgo pkg-config: librsvg-2.0 cairo
// #cgo LDFLAGS: -lm
// #include "librsvg_wrapper.h"
import "C"

import (
	// "bytes"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io"
	"strings"

	// "strings"
	"unsafe"
)

// Render a simple a svg to an image.Image
func Decode(reader io.Reader, w, h int) (image.Image, error) {
	buf, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading data: %v", err)
	}

	return renderSVG(buf, w, h, "")
}

// For tinting symbolic icons
func DecodeWithColor(reader io.Reader, w, h int, c color.Color) (image.Image, error) {
	r, g, b, _ := c.RGBA()
	hexColor := fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))

	css := fmt.Sprintf(`
      path {
        fill: %s !important;
      }
      svg { color: %s; }
    `, hexColor, hexColor)

	buf, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading data: %v", err)
	}
	reader = bytes.NewReader([]byte(strings.NewReplacer(
		"fill:#000000", "fill:"+hexColor,
		"fill=\"#000000", "fill=\""+hexColor,
	).Replace(string(buf))))
	return DecodeWithCSS(reader, w, h, css)
}

// For applying custom CSS
func DecodeWithCSS(reader io.Reader, w, h int, css string) (image.Image, error) {
	buf, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading data: %v", err)
	}

	return renderSVG(buf, w, h, css)
}

func renderSVG(svgData []byte, width, height int, css string) (image.Image, error) {
	cSvgData := C.CBytes(svgData)
	defer C.free(cSvgData)

	var cCSS *C.char = nil
	if css != "" {
		cCSS = C.CString(css)
		defer C.free(unsafe.Pointer(cCSS))
	}

	result := C.render_svg_css((*C.char)(cSvgData), C.int(len(svgData)), cCSS, C.int(width), C.int(height))

	if result == nil {
		return nil, fmt.Errorf("failed to render SVG")
	}
	defer C.free_render_result(result)

	rWidth := int(result.width)
	rHeight := int(result.height)
	rStride := int(result.stride)

	img := image.NewRGBA(image.Rect(0, 0, rWidth, rHeight))

	dataPtr := unsafe.Pointer(result.data)
	dataSlice := unsafe.Slice((*byte)(dataPtr), rStride*rHeight)

	copy(img.Pix, dataSlice)

	return img, nil
}
