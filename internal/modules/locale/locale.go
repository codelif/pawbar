package locale

import(
	"os"
	"fmt"
	"strings"

	"github.com/codelif/pawbar/internal/modules"
)

func New() modules.Module{
	return &LocaleModule{}
}

type LocaleModule struct{
		receive chan bool
		send    chan modules.Event
}

func (l *LocaleModule) Dependencies() []string {
	return nil
}

func (l *LocaleModule) splitLocales(locales string) []string {
	return strings.Split(locales, ":")
}

func (l *LocaleModule) getLangFromEnv() string {
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

func (l *LocaleModule)  getUnixLocales() []string {
	locale := getLangFromEnv()
	if locale == "C" || locale == "POSIX" || len(locale) == 0 {
		return nil
	}

	return splitLocales(locale)
}

func (l *LocaleModule) GetLocale() (string, error) {
	unixLocales := getUnixLocales()
	if unixLocales == nil {
		return "", nil
	}

	language, region := splitLocale(unixLocales[0])
	locale := language
	if len(region) > 0 {
		locale = strings.Join([]string{language, region}, "-")
	}

	return locale, nil
}

func (c *LocaleModule) Run() (<-chan bool, chan<- modules.Event, error) {
	l.receive = make(chan bool)
	l.send = make(chan modules.Event)

	go func() {
		t := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-t.C:
				c.receive <- true
			case e := <-c.send:
			}
		}
	}()

	return l.receive, l.send, nil
}

func (l *LocaleModule) Render() []modules.EventCell {
	rstring, err  := l.GetLocale()
	if err !=nil {
		return ""
	}
	r := make([]modules.EventCell, len(rstring))
	for i, ch := range rstring {
		r[i] = modules.EventCell{C: ch, Style: modules.DEFAULT, Metadata: "", Mod: c}
	}

	return r
}

func (l *LocaleModule) Channels() (<-chan bool, chan<- modules.Event) {
	return l.receive, l.send
}

func (l *LocaleModule) Name() string {
	return "locale"
}
