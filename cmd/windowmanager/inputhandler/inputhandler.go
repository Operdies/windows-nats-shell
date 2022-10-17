package inputhandler

import (
	"github.com/operdies/windows-nats-shell/pkg/input"
	"github.com/operdies/windows-nats-shell/pkg/input/keyboard"
	"github.com/operdies/windows-nats-shell/pkg/input/mouse"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/windows"
	"github.com/operdies/windows-nats-shell/pkg/winapi"
	"github.com/operdies/windows-nats-shell/pkg/winapi/windowmanager"
	"github.com/operdies/windows-nats-shell/pkg/winapi/wintypes"
)

var actionKey = input.VK_MAP["nullkey"]

type mouseButtonState struct {
	lButtonDown bool
	rButtonDown bool
}

type InputHandler struct {
	keyMods         map[input.VKEY]bool
	mouseMods       mouseButtonState
	resizeEventInfo resizeEventInfo
	dragEventInfo   dragEventInfo
}

type resizeEventInfo struct {
	// The mouse event which triggered the resize
	trigger mouse.MouseEventInfo
	// Is the event currently active
	active bool
	// The resizing handle
	resizeDirection windows.WindowCardinals
	// The window handle
	handle wintypes.HWND
	// The initial configuration of the window
	startPosition wintypes.RECT
}

type dragEventInfo struct {
	// The mouse event which triggered the move
	trigger mouse.MouseEventInfo
	// Is the event currently active
	active bool
	// The starting position of the window
	startPosition wintypes.RECT
	// The window handle
	handle wintypes.HWND
}

func CreateInputHandler() *InputHandler {
	var handler InputHandler
	handler.keyMods = map[input.VKEY]bool{}

	return &handler
}

type direction int

const (
	Next     direction = 1
	Previous           = 2
)

func SelectSibling(siblingOf wintypes.HWND, dir direction) {
	var focus wintypes.HWND
	prev, next := windowmanager.GetSiblings(siblingOf)

	if dir == Next {
		focus = next
	} else {
		focus = prev
	}
	winapi.SuperFocusStealer(focus)
}

func (k *InputHandler) OnKeyboardInput(kei keyboard.KeyboardEventInfo) bool {
	// We can't differentiate presses and holds from the event info.
	// But we know it's a hold if the key is already mapped in activeMods
	keyDown := kei.TransitionState == false

	// Update the modifier state if the key is a modifier
	vkey := input.VKEY(kei.VirtualKeyCode)
	// Clear the keymap when escape is pressed
	if vkey == input.VK_ESCAPE {
		k.keyMods = map[input.VKEY]bool{}
		return false
	}

	k.keyMods[vkey] = keyDown
	return vkey == actionKey
}

func (h *InputHandler) resizeStart(mei mouse.MouseEventInfo) {
	subject := winapi.WindowFromPoint(mei.Point)
	if subject == 0 {
		return
	}

	rect := winapi.GetWindowRect(subject)
	corner := windows.GetNearestCardinal(mei.Point, rect)

	h.resizeEventInfo = resizeEventInfo{
		active:          true,
		trigger:         mei,
		resizeDirection: corner,
		handle:          subject,
		startPosition:   rect,
	}
}
func applyResize(h *InputHandler, p wintypes.POINT) {
	delta := h.resizeEventInfo.trigger.Point.Sub(p)
	d := h.resizeEventInfo.resizeDirection
	r := h.resizeEventInfo.startPosition
	if d&windows.Top > 0 {
		r.Top -= int32(delta.Y)
	}
	if d&windows.Bottom > 0 {
		r.Bottom -= int32(delta.Y)
	}
	if d&windows.Left > 0 {
		r.Left -= int32(delta.X)
	}
	if d&windows.Right > 0 {
		r.Right -= int32(delta.X)
	}
	windowmanager.SetWindowRect(h.resizeEventInfo.handle, r)
}
func (h *InputHandler) resizing(mei mouse.MouseEventInfo) {
	applyResize(h, mei.Point)
}
func (h *InputHandler) resizeEnd(mei mouse.MouseEventInfo) {
	applyResize(h, mei.Point)
	h.resizeEventInfo.active = false
}

func (h *InputHandler) dragStart(mei mouse.MouseEventInfo) {
	subject := winapi.WindowFromPoint(mei.Point)
	if subject == 0 {
		return
	}
	rect := winapi.GetWindowRect(subject)
	h.dragEventInfo.active = true
	h.dragEventInfo.trigger = mei
	h.dragEventInfo = dragEventInfo{
		active:        true,
		trigger:       mei,
		handle:        subject,
		startPosition: rect,
	}
}

func applyMove(h *InputHandler, p wintypes.POINT) {
	// delta := h.dragEventInfo.trigger.Point.Sub(p)
	// r := h.dragEventInfo.startPosition.Transform(-int32(delta.X), -int32(delta.Y))
	// windowmanager.SetWindowRect(h.dragEventInfo.handle, r)

	delta := h.dragEventInfo.trigger.Point.Sub(p)
	startPoint := wintypes.POINT{X: wintypes.LONG(h.dragEventInfo.startPosition.Left), Y: wintypes.LONG(h.dragEventInfo.startPosition.Top)}
	target := startPoint.Sub(delta)
	windowmanager.MoveWindow(h.dragEventInfo.handle, target)
}

func (h *InputHandler) dragging(mei mouse.MouseEventInfo) {
	applyMove(h, mei.Point)
}
func (h *InputHandler) dragEnd(mei mouse.MouseEventInfo) {
	applyMove(h, mei.Point)
	h.dragEventInfo.active = false
}

func (h *InputHandler) OnMouseInput(mei mouse.MouseEventInfo) bool {
	switch mei.Action {
	case mouse.LBUTTONDOWN:
		h.mouseMods.lButtonDown = true
		if state, ok := h.keyMods[actionKey]; ok && state && !h.resizeEventInfo.active {
			h.dragStart(mei)
			return true
		}
	case mouse.LBUTTONUP:
		h.mouseMods.lButtonDown = false
		if h.dragEventInfo.active {
			h.dragEnd(mei)
			return true
		}
	case mouse.RBUTTONDOWN:
		if state, ok := h.keyMods[actionKey]; ok && state && !h.dragEventInfo.active {
			h.resizeStart(mei)
			return true
		}
		h.mouseMods.rButtonDown = true
	case mouse.RBUTTONUP:
		h.mouseMods.rButtonDown = false
		if h.resizeEventInfo.active {
			h.resizeEnd(mei)
			return true
		}
	case mouse.MOUSEMOVE:
		if h.dragEventInfo.active {
			h.dragging(mei)
		} else if h.resizeEventInfo.active {
			h.resizing(mei)
		}
	case mouse.VMOUSEWHEEL:
		if state, ok := h.keyMods[actionKey]; ok && state {
			hwnd := winapi.GetForegroundWindow()
			var d direction
			if mei.WheelDelta > 0 {
				d = Next
			} else {
				d = Previous
			}
			SelectSibling(hwnd, d)
			return true
		}
	}
	return false
}
