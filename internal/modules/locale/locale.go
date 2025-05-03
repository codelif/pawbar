package locale

import (
	"os"
	"strings"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/codelif/pawbar/internal/modules"
)

func New() modules.Module {
	return &LocaleModule{}
}

type LocaleModule struct {
	receive chan bool
	send    chan modules.Event
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

	go func() {
		t := time.NewTicker(7 * time.Second)
		for {
			select {
			case <-t.C:
				mod.receive <- true
			case <-mod.send:
			}
		}
	}()

	return mod.receive, mod.send, nil
}

func (mod *LocaleModule) Render() []modules.EventCell {
	rstring, err := mod.GetLocale()
	if err != nil {
		return nil
	}

	rch := vaxis.Characters(rstring)
	r := make([]modules.EventCell, len(rch))
	for i, ch := range rch {
		r[i] = modules.EventCell{C: vaxis.Cell{Character: ch}, Mod: mod}
	}

	return r
}

func (mod *LocaleModule) Channels() (<-chan bool, chan<- modules.Event) {
	return mod.receive, mod.send
}

func (mod *LocaleModule) Name() string {
	return "locale"
}
