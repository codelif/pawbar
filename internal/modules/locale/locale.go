package locale

import (
	"bytes"
	"os"
	"strings"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/nekorg/pawbar/internal/config"
	"github.com/nekorg/pawbar/internal/modules"
)

func New() modules.Module {
	return &LocaleModule{}
}

type LocaleModule struct {
	receive chan bool
	send    chan modules.Event

	opts        Options
	initialOpts Options

	currentTickerInterval time.Duration
	ticker                *time.Ticker
}

func (mod *LocaleModule) Dependencies() []string {
	return nil
}

func (mod *LocaleModule) splitLocale(locale string) (string, string) {
	formattedLocale, _, _ := strings.Cut(locale, ".")
	formattedLocale = strings.ReplaceAll(formattedLocale, "-", "_")
	language, territory, _ := strings.Cut(formattedLocale, "_")
	return language, territory
}

func (mod *LocaleModule) splitLocales(locales string) []string {
	return strings.Split(locales, ":")
}

func (mod *LocaleModule) getLangFromEnv() string {
	locale := ""
	for _, env := range [...]string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		locale = os.Getenv(env)
		if len(locale) > 0 {
			break
		}
	}

	if locale == "C" || locale == "POSIX" {
		return locale
	}
	languages := os.Getenv("LANGUAGE")
	if len(languages) > 0 {
		return languages
	}

	return locale
}

func (mod *LocaleModule) getUnixLocales() []string {
	locale := mod.getLangFromEnv()
	if locale == "C" || locale == "POSIX" || len(locale) == 0 {
		return nil
	}

	return mod.splitLocales(locale)
}

func (mod *LocaleModule) GetLocale() (string, error) {
	unixLocales := mod.getUnixLocales()
	if unixLocales == nil {
		return "", nil
	}

	language, region := mod.splitLocale(unixLocales[0])
	locale := language
	if len(region) > 0 {
		locale = strings.Join([]string{language, region}, "-")
	}

	return locale, nil
}

func (mod *LocaleModule) Run() (<-chan bool, chan<- modules.Event, error) {
	mod.receive = make(chan bool)
	mod.send = make(chan modules.Event)
	mod.initialOpts = mod.opts

	go func() {
		mod.currentTickerInterval = mod.opts.Tick.Go()
		mod.ticker = time.NewTicker(mod.currentTickerInterval)
		defer mod.ticker.Stop()
		for {
			select {
			case <-mod.ticker.C:
				mod.receive <- true
			case e := <-mod.send:
				switch ev := e.VaxisEvent.(type) {
				case vaxis.Mouse:
					if ev.EventType != vaxis.EventPress {
						break
					}
					btn := config.ButtonName(ev)
					if mod.opts.OnClick.Dispatch(btn, &mod.initialOpts, &mod.opts) {
						mod.receive <- true
					}
					mod.ensureTickInterval()

				case modules.FocusIn:
					if mod.opts.OnClick.HoverIn(&mod.opts) {
						mod.receive <- true
					}
					mod.ensureTickInterval()

				case modules.FocusOut:
					if mod.opts.OnClick.HoverOut(&mod.opts) {
						mod.receive <- true
					}
					mod.ensureTickInterval()
				}
			}
		}
	}()

	return mod.receive, mod.send, nil
}

func (mod *LocaleModule) ensureTickInterval() {
	if mod.opts.Tick.Go() != mod.currentTickerInterval {
		mod.currentTickerInterval = mod.opts.Tick.Go()
		mod.ticker.Reset(mod.currentTickerInterval)
	}
}

func (mod *LocaleModule) Render() []modules.EventCell {
	locale, err := mod.GetLocale()
	if err != nil {
		return nil
	}

	style := vaxis.Style{
		Foreground: mod.opts.Fg.Go(),
		Background: mod.opts.Bg.Go(),
	}

	data := struct {
		Locale string
	}{
		Locale: locale,
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

func (mod *LocaleModule) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *LocaleModule) Name() string {
	return "locale"
}
