package all

import (
	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
	_ "github.com/codelif/pawbar/internal/modules/backlight"
	_ "github.com/codelif/pawbar/internal/modules/battery"
	_ "github.com/codelif/pawbar/internal/modules/bluetooth"
	_ "github.com/codelif/pawbar/internal/modules/clock"
	_ "github.com/codelif/pawbar/internal/modules/cpu"
	_ "github.com/codelif/pawbar/internal/modules/disk"
	_ "github.com/codelif/pawbar/internal/modules/locale"
	_ "github.com/codelif/pawbar/internal/modules/ram"
	_ "github.com/codelif/pawbar/internal/modules/title"
	_ "github.com/codelif/pawbar/internal/modules/volume"
	_ "github.com/codelif/pawbar/internal/modules/wifi"
	_ "github.com/codelif/pawbar/internal/modules/ws"
	"gopkg.in/yaml.v3"
)

func init() {
	config.Register("sep", func(n *yaml.Node) (modules.Module, error) {
		return modules.NewStaticModule(
			"sep",
			[]modules.EventCell{
				{C: modules.ECSPACE.C},
				{C: vaxis.Cell{
					Character: vaxis.Character{
						Grapheme: "│",
						Width:    1,
					},
				}},
				{C: modules.ECSPACE.C},
			}, nil,
		), nil
	})

	config.Register("space", func(raw *yaml.Node) (modules.Module, error) {
		return modules.NewStaticModule(
			"space",
			[]modules.EventCell{
				{C: modules.ECSPACE.C},
			},
			nil,
		), nil
	})
}
