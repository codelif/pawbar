package main

import (
	"image"
	"image/png"
	"os"

	"github.com/codelif/pawbar/pkg/gorsvg"
	"golang.org/x/image/colornames"
)

func main() {
	f, err := os.Open("/usr/share/icons/Adwaita/symbolic/devices/bluetooth-symbolic.svg")
	if err != nil {
		panic(err)
	}

	var img image.Image
	if len(os.Args) > 1 {
		img, err = gorsvg.DecodeWithColor(f, 48, 48, colornames.Indigo)
		if err != nil {
			panic(err)
		}
	} else {
		img, err = gorsvg.Decode(f, 48, 48)
		if err != nil {
			panic(err)
		}
	}

	png.Encode(os.Stdout, img)
}
