package inputhandler

import (
	"fmt"

	"github.com/operdies/windows-nats-shell/pkg/input"
	"github.com/operdies/windows-nats-shell/pkg/input/keyboard"
	"github.com/operdies/windows-nats-shell/pkg/input/mouse"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/windows"
	"github.com/operdies/windows-nats-shell/pkg/winapi"
	"github.com/operdies/windows-nats-shell/pkg/winapi/windowmanager"
	"github.com/operdies/windows-nats-shell/pkg/winapi/wintypes"
)

var actionKey = input.VK_MAP["nullkey"]

type InputHandler struct {
	keyMods   map[input.VKEY]bool
	eventInfo eventInfo
}

type eventInfo struct {
	// The mouse event which triggered the resize
	trigger mouse.MouseEventInfo
	// are we currently dragging
	dragging bool
	// are we currently resizing
	resizing bool
	// The resizing handle
	resizeDirection windows.WindowCardinals
	// The window handle
	handle wintypes.HWND
	// The initial configuration of the window
	startPosition wintypes.RECT
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

func getRootOwnerAtPoint(mei mouse.MouseEventInfo) wintypes.HWND {
	subject := winapi.WindowFromPoint(mei.Point)
	if subject == 0 {
		return 0
	}
	rootOwner := winapi.GetAncestor(subject, wintypes.GA_ROOTOWNER)
	if rootOwner != 0 {
		subject = rootOwner
	}

	return subject
}

func (h *InputHandler) resizeStart(mei mouse.MouseEventInfo) {
	subject := getRootOwnerAtPoint(mei)

	rect := winapi.GetWindowRect(subject)
	corner := windows.GetNearestCardinal(mei.Point, rect)

	h.eventInfo = eventInfo{
		resizing:        true,
		trigger:         mei,
		resizeDirection: corner,
		handle:          subject,
		startPosition:   rect,
	}
}
func applyResize(h *InputHandler, p wintypes.POINT) {
	delta := h.eventInfo.trigger.Point.Sub(p)
	d := h.eventInfo.resizeDirection
	r := h.eventInfo.startPosition
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
	windowmanager.SetWindowRect(h.eventInfo.handle, r)
}
func (h *InputHandler) resizing(mei mouse.MouseEventInfo) {
	applyResize(h, mei.Point)
}
func (h *InputHandler) resizeEnd(mei mouse.MouseEventInfo) {
	h.resizing(mei)
	h.eventInfo.resizing = false
}

func (h *InputHandler) dragStart(mei mouse.MouseEventInfo) {
	subject := getRootOwnerAtPoint(mei)
	if subject == 0 {
		return
	}
	rect := winapi.GetWindowRect(subject)
	h.eventInfo = eventInfo{
		dragging:      true,
		trigger:       mei,
		handle:        subject,
		startPosition: rect,
	}
}

func applyMove(h *InputHandler, p wintypes.POINT) {
	// delta := h.resizeEventInfo.trigger.Point.Sub(p)
	// r := h.resizeEventInfo.startPosition.Transform(-int32(delta.X), -int32(delta.Y))
	// windowmanager.SetWindowRect(h.resizeEventInfo.handle, r)

	delta := h.eventInfo.trigger.Point.Sub(p)
	startPoint := wintypes.POINT{X: wintypes.LONG(h.eventInfo.startPosition.Left), Y: wintypes.LONG(h.eventInfo.startPosition.Top)}
	target := startPoint.Sub(delta)
	windowmanager.MoveWindow(h.eventInfo.handle, target)
}

func (h *InputHandler) dragging(mei mouse.MouseEventInfo) {
	applyMove(h, mei.Point)
}
func (h *InputHandler) dragEnd(mei mouse.MouseEventInfo) {
	h.dragging(mei)
	h.eventInfo.dragging = false
}

func (h *InputHandler) printMouseover(mei mouse.MouseEventInfo) {
	subject := winapi.WindowFromPoint(mei.Point)
	rootOwner := winapi.GetAncestor(subject, wintypes.GA_ROOTOWNER)
	windowTitle, _ := windowmanager.GetWindowTextEasy(subject)
	fmt.Printf("subject %v) %v\n", subject, windowTitle)
	windowTitle, _ = windowmanager.GetWindowTextEasy(rootOwner)
	fmt.Printf("subject owner %v) %v\n", rootOwner, windowTitle)
}

func (h *InputHandler) OnMouseInput(mei mouse.MouseEventInfo) bool {
	switch mei.Action {
	case mouse.LBUTTONDOWN:
		if state, ok := h.keyMods[actionKey]; ok && state && !h.eventInfo.resizing {
			h.dragStart(mei)
			return true
		}
	case mouse.LBUTTONUP:
		if h.eventInfo.dragging {
			h.dragEnd(mei)
			return true
		}
	case mouse.RBUTTONDOWN:
		if state, ok := h.keyMods[actionKey]; ok && state && !h.eventInfo.dragging {
			h.resizeStart(mei)
			return true
		}
	case mouse.RBUTTONUP:
		if h.eventInfo.resizing {
			h.resizeEnd(mei)
			return true
		}
	case mouse.MOUSEMOVE:
		if h.eventInfo.dragging {
			h.dragging(mei)
		} else if h.eventInfo.resizing {
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
