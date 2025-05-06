package mouse

import "git.sr.ht/~rockorager/vaxis"

func ButtonName(ev vaxis.Mouse) string {
	switch ev.Button {
	case vaxis.MouseLeftButton:
		return "left"
	case vaxis.MouseMiddleButton:
		return "middle"
	case vaxis.MouseRightButton:
		return "right"
	case vaxis.MouseWheelUp:
		return "wheel-up"
	case vaxis.MouseWheelDown:
		return "wheel-down"
	case 66:
		return "wheel-left"
	case 67:
		return "wheel-right"
	}

	return ""
}

func IsPress(ev vaxis.Mouse) bool  { return ev.EventType == vaxis.EventPress }
func IsScroll(ev vaxis.Mouse) bool { return ev.Button >= vaxis.MouseWheelUp && ev.Button <= 67 }
