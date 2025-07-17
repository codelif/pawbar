package menu

import "github.com/codelif/xdgicons"

const (
	ItemStandard  string = "standard"
	ItemSeparator string = "separator"
)

const (
	ToggleNone      string = ""
	ToggleCheckMark string = "checkmark"
	ToggleRadio     string = "radio"
)

const (
	StateOff           int = 0
	StateOn            int = 1
	StateIndeterminate int = -1 // can be anything other than 0&1 but set to default value
)

type MessageType int

const (
	MsgMenuUpdate MessageType = iota
	MsgMouseUpdate
	MsgItemClicked
	MsgItemHovered
	MsgSubmenuRequested
	MsgSubmenuCancelRequested
)

type MessagePayload struct {
	Menu   []Item
	ItemId int32
	MouseX int
	MouseY int
	PixelX int
	PixelY int
}

type Message struct {
	Type    MessageType
	Payload MessagePayload
}

type Label struct {
	Display     string
	AccessKey   rune
	AccessIndex int
	Found       bool
}

type Item struct {
	Id          int32
	Type        string
	Label       Label
	Enabled     bool
	Visible     bool
	IconName    string
	Icon        xdgicons.Icon // custom property for speeding up icon lookup (valid if IconName exists)
	IconData    []byte
	Shortcut    [][]string
	ToggleType  string
	ToggleState int32
	HasChildren bool
}

func ParseLabel(label string) Label {
	runes := []rune(label)
	n := len(runes)

	var output []rune
	outPos := 0
	var result Label

	for i := 0; i < n; {
		if runes[i] == '_' {
			if i+1 < n && runes[i+1] == '_' {
				output = append(output, '_')
				outPos++
				i += 2
			} else {
				if !result.Found && i+1 < n {
					result.Found = true
					result.AccessKey = runes[i+1]
					result.AccessIndex = outPos
					output = append(output, runes[i+1])
					outPos++
					i += 2
				} else {
					i++
				}
			}
		} else {
			output = append(output, runes[i])
			outPos++
			i++
		}
	}

	result.Display = string(output)
	return result
}
