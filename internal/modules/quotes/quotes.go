package quotes

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/config"
	"github.com/codelif/pawbar/internal/modules"
)

type Quotes struct {
	receive chan bool
	send    chan modules.Event
	quote   string

	opts        Options
	initialOpts Options
}

func (mod *Quotes) Dependencies() []string {
	return []string{}
}

func (mod *Quotes) Name() string {
	return "quotes"
}

func New() modules.Module {
	return &Quotes{}
}

func (mod *Quotes) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *Quotes) pickQuote() (string, error) {
	cmd := exec.Command("bash", "-c", "~/.config/pawbar/pickQuotes.sh")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Ran into error extracting quote")
	}

	result := string(output)
	return strings.TrimSpace(result), nil
}

func (mod *Quotes) Run() (<-chan bool, chan<- modules.Event, error) {
	mod.receive = make(chan bool)
	mod.send = make(chan modules.Event)
	mod.initialOpts = mod.opts

	quote, err := mod.pickQuote()
	if err != nil {
		return nil, nil, err
	}
	mod.quote = quote

	go func() {
		for {
			select {
			case e := <-mod.send:
				switch ev := e.VaxisEvent.(type) {
				case vaxis.Mouse:
					if ev.EventType != vaxis.EventPress && ev.EventType != vaxis.EventRelease {
						break
					}
					btn := config.ButtonName(ev)

					if btn == "left" {
						quote, err := mod.pickQuote()
						if err != nil {
							continue
						}
						mod.quote = quote
						mod.receive <- true

					}
					if mod.opts.OnClick.Dispatch(btn, &mod.initialOpts, &mod.opts) {
						mod.receive <- true
					}

				case modules.FocusIn:
					if mod.opts.OnClick.HoverIn(&mod.opts) {
						mod.receive <- true
					}

				case modules.FocusOut:
					if mod.opts.OnClick.HoverOut(&mod.opts) {
						mod.receive <- true
					}
				}
			}
		}
	}()

	return mod.receive, mod.send, nil
}

func (mod *Quotes) Render() []modules.EventCell {
	style := vaxis.Style{
		Foreground: mod.opts.Fg.Go(),
		Background: mod.opts.Bg.Go(),
	}

	data := struct {
		Quote string
	}{
		Quote: mod.quote,
	}

	var buf bytes.Buffer
	_ = mod.opts.Format.Execute(&buf, data)

	rch := vaxis.Characters(buf.String())
	r := make([]modules.EventCell, len(rch))

	for i, ch := range rch {
		r[i] = modules.EventCell{C: vaxis.Cell{Character: ch, Style: style}, Mod: mod, MouseShape: mod.opts.Cursor.Go()}
	}
	return r
}
