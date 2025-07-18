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

	// Only start timer for submenu items that don't already have an active timer
	if item.HasChildren && item.Enabled {
		// Only set new timer if this is a different item or no timer is active
		if h.state.hoverItemId != item.Id {
			// If we had a different submenu item, cancel it first
			if h.state.hoverItemId != 0 {
				h.handleSubmenuCancel()
			}

			h.state.hoverItemId = item.Id
			capturedItemId := item.Id
			capturedMouseY := h.state.mouseY

			h.state.hoverTimer = time.AfterFunc(hoverActivationTimeout, func() {
				// Only proceed if we're still hovering the same item at same position
				if h.state.hoverItemId == capturedItemId &&
					h.state.mouseY == capturedMouseY &&
					h.state.mouseOnSurface {
					h.sendMessage(menu.MsgSubmenuRequested, item.Id,
						h.state.mousePixelX, h.state.mousePixelY)
				}
			})
		}
		// If it's the same item, do nothing - keep existing state
	} else {
		// Moving to non-submenu item - cancel any active submenu
		if h.state.hoverItemId != 0 {
			h.handleSubmenuCancel()
		}
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
	// If mouse hasn't actually moved to a different row, ignore
	if h.state.mouseY == row && h.state.mouseOnSurface {
		return
	}

	prevMouseY := h.state.mouseY
	h.state.mouseOnSurface = true
	h.state.mouseY = row

	// Cancel any pending timer (but keep hoverItemId for submenu tracking)
	h.state.cancelHoverTimer()

	// If we moved to a different item, handle the transition
	if prevMouseY != row {
		// Handle hover for valid items
		if h.state.isSelectableItem(h.state.mouseY) {
			h.handleItemHover(&h.state.items[h.state.mouseY])
		} else {
			// Hovering over separator or invalid area - cancel any submenu
			if h.state.hoverItemId != 0 {
				h.handleSubmenuCancel()
			}
			h.state.lastMouseY = -1
		}
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
