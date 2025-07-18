package tui

import (
	"time"

	"github.com/codelif/pawbar/pkg/dbusmenukitty/menu"
	"github.com/fxamacker/cbor/v2"
)
type MessageHandler struct {
	encoder *cbor.Encoder
	state   *MenuState
}

func NewMessageHandler(encoder *cbor.Encoder, state *MenuState) *MessageHandler {
	return &MessageHandler{
		encoder: encoder,
		state:   state,
	}
}

func (h *MessageHandler) sendMessage(msgType menu.MessageType, itemId int32, pixelX, pixelY int) {
	msg := menu.Message{
		Type: msgType,
		Payload: menu.MessagePayload{
			ItemId: itemId,
			PixelX: pixelX,
			PixelY: pixelY,
		},
	}
	h.encoder.Encode(msg)
}

func (h *MessageHandler) handleItemHover(item *menu.Item) {
	h.sendMessage(menu.MsgItemHovered, item.Id, 0, 0)
	h.state.lastMouseY = h.state.mouseY
	
	// Start hover timer for submenu items
	if item.HasChildren && item.Enabled && item.Id != h.state.hoverItemId {
		h.state.hoverItemId = item.Id
		capturedItemId := item.Id

		h.state.hoverTimer = time.AfterFunc(hoverActivationTimeout, func() {
			if h.state.hoverItemId == capturedItemId {
				h.sendMessage(menu.MsgSubmenuRequested, item.Id, 
					h.state.mousePixelX, h.state.mousePixelY)
			}
		})
	}
}

func (h *MessageHandler) handleSubmenuCancel() {
	if h.state.hoverItemId != 0 {
		h.sendMessage(menu.MsgSubmenuCancelRequested, h.state.hoverItemId, 0, 0)
		h.state.hoverItemId = 0
	}
}

func (h *MessageHandler) handleItemClick(item *menu.Item) {
	h.sendMessage(menu.MsgItemClicked, item.Id, 0, 0)
}

func (h *MessageHandler) handleMouseMotion(row int) {
	if h.state.mouseY == row && h.state.mouseOnSurface {
		return
	}
	
	h.state.mouseOnSurface = true
	h.state.mouseY = row
	h.state.cancelHoverTimer()
	
	// Cancel previous submenu if switching items
	if h.state.lastMouseY != -1 && h.state.lastMouseY != h.state.mouseY && h.state.hoverItemId != 0 {
		h.handleSubmenuCancel()
	}
	
	// Handle hover for valid items
	if h.state.isSelectableItem(h.state.mouseY) {
		h.handleItemHover(&h.state.items[h.state.mouseY])
	} else {
		// Hovering over separator or invalid area
		h.handleSubmenuCancel()
		h.state.lastMouseY = -1
	}
}

func (h *MessageHandler) handleKeyNavigation(keyPressed bool) {
	if !keyPressed || !h.state.isValidItemIndex(h.state.mouseY) {
		return
	}
	
	h.state.cancelHoverTimer()
	currentItem := &h.state.items[h.state.mouseY]
	h.handleItemHover(currentItem)
}

