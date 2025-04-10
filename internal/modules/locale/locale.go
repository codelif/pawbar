package locale

import(
	"os"
	"strings"
	"time"

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

func (l *LocaleModule) splitLocale(locale string) (string, string) {
	formattedLocale, _, _ := strings.Cut(locale, ".")
	formattedLocale = strings.ReplaceAll(formattedLocale, "-", "_")
	language, territory, _ := strings.Cut(formattedLocale, "_")
	return language, territory
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
	locale := l.getLangFromEnv()
	if locale == "C" || locale == "POSIX" || len(locale) == 0 {
		return nil
	}

	return l.splitLocales(locale)
}

func (l *LocaleModule) GetLocale() (string, error) {
	unixLocales := l.getUnixLocales()
	if unixLocales == nil {
		return "", nil
	}

	language, region := l.splitLocale(unixLocales[0])
	locale := language
	if len(region) > 0 {
		locale = strings.Join([]string{language, region}, "-")
	}

	return locale, nil
}

func (l *LocaleModule) Run() (<-chan bool, chan<- modules.Event, error) {
	l.receive = make(chan bool)
	l.send = make(chan modules.Event)

	go func() {
		t := time.NewTicker(7 * time.Second)
		for {
			select {
			case <-t.C:
				l.receive <- true
			case <-l.send:
			}
		}
	}()

	return l.receive, l.send, nil
}

func (l *LocaleModule) Render() []modules.EventCell {
	rstring, err  := l.GetLocale()
	if err !=nil {
		return nil
	}
	r := make([]modules.EventCell, len(rstring))
	for i, ch := range rstring {
		r[i] = modules.EventCell{C: ch, Style: modules.DEFAULT, Metadata: "", Mod: l}
	}

	return r
}

func (l *LocaleModule) Channels() (<-chan bool, chan<- modules.Event) {
	return l.receive, l.send
}

func (l *LocaleModule) Name() string {
	return "locale"
}
