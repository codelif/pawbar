package mouse

import (
	"os/exec"
	"reflect"

	"git.sr.ht/~rockorager/vaxis"
)

type clickProvider[C any] interface {
	GetOnClicks() Map[C]
	GetCursor() string
}

type Mixin[C clickProvider[C]] struct {
	Handlers Map[C]
	Apply    func(*C) bool
	Cursor   vaxis.MouseShape
}

func (m *Mixin[C]) Handle(ev vaxis.Mouse) (changed bool) {
	if ev.EventType != vaxis.EventPress {
		return
	}
	btn := ButtonName(ev)

	h, ok := m.Handlers[btn]
	if !ok {
		return
	}

	switch {
	case h.Config != nil:
		changed = m.Apply(h.Config)
		if c := (*h.Config).GetCursor(); c != "" {
			newShape := ParseCursor(c)
			if newShape != m.Cursor {
				m.Cursor = newShape
				changed = true
			}
		}
	case len(h.Run) > 0:
		changed = true
		go exec.Command(h.Run[0], h.Run[1:]...).Start()
	case h.Notify != "":
		changed = true
		go exec.Command("notify-send", h.Notify).Run()
	}

	if h.Config != nil {
		for btn, handler := range (*h.Config).GetOnClicks() {
			m.Handlers[btn] = handler
		}
		// m.Handlers = (*h.Config).GetOnClicks()
	}

	return

}

func Overlay[T any](dst *T, src *T) (changed bool) {
	changed = overlay(reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem())
	return changed
}

func overlay(dst, src reflect.Value) (changed bool) {
	for i := 0; i < dst.NumField(); i++ {
		fDst, fSrc := dst.Field(i), src.Field(i)

		switch fSrc.Kind() {
		case reflect.String:
			if s := fSrc.String(); s != "" && fDst.String() != s {
				fDst.SetString(s)
				changed = true
			}

		case reflect.Int, reflect.Int64:
			if v := fSrc.Int(); v != 0 && fDst.Int() != v {
				fDst.SetInt(v)
				changed = true
			}

		case reflect.Uint, reflect.Uint64:
			if v := fSrc.Uint(); v != 0 && fDst.Uint() != v {
				fDst.SetUint(v)
				changed = true
			}

		case reflect.Bool:
			if b := fSrc.Bool(); b && fDst.Bool() != b {
				fDst.SetBool(b)
				changed = true
			}
		case reflect.Struct:
			if overlay(fDst, fSrc) {
				changed = true
			}
		}
	}

	return
}
