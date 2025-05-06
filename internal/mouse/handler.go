package mouse

import "git.sr.ht/~rockorager/vaxis"

type Handler[C any] struct {
	Config *C       `yaml:",inline"`
	Run    []string `yaml:"run"`
	Notify string   `yaml:"notify"`
}

type Map[C any] map[string]Handler[C]

func (h Handler[C]) IsEmpty() bool {
	return h.Config == nil && len(h.Run) == 0 && h.Notify == ""
}

func ParseCursor(s string) vaxis.MouseShape {
	switch s {
	case "pointer":
		return vaxis.MouseShapeClickable
	case "default", "":
		return vaxis.MouseShapeDefault
	case "text":
		return vaxis.MouseShapeTextInput
	default:
		return vaxis.MouseShapeDefault
	}
}
