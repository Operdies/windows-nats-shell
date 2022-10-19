package mouse

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/operdies/windows-nats-shell/pkg/input"
	"github.com/operdies/windows-nats-shell/pkg/winapi"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"
)

type _MSLLHOOKSTRUCT struct {
	pt          wintypes.POINT
	mouseData   wintypes.DWORD
	flags       wintypes.DWORD
	time        wintypes.DWORD
	dwExtraInfo wintypes.LONG_PTR
}

type MouseAction = uint32

const (
	MOUSEMOVE   MouseAction = 0x200
	LBUTTONDOWN             = 0x201
	LBUTTONUP               = 0x202
	RBUTTONDOWN             = 0x204
	RBUTTONUP               = 0x205
	VMOUSEWHEEL             = 0x20A
	HMOUSEWHEEL             = 0x20E
)

type MouseEventInfo struct {
	// The mouse action being triggered
	Action MouseAction
	// The point in per-monitor aware screen coordinates the action was triggered
	Point wintypes.POINT
	// A positive value indicates that the wheel was rotated forward, away from the user; a negative value indicates that the wheel was rotated backward, toward the user. One wheel click is defined as WHEEL_DELTA, which is 120.
	WheelDelta int16
	// The time stamp for this message.
	Time wintypes.DWORD
}

func Words(dword wintypes.DWORD) (lower, higher int16) {
	return int16(dword & 0xFF),
		int16(dword >> 16)
}

func WhMouseEvent(wParam wintypes.WPARAM, info _MSLLHOOKSTRUCT) MouseEventInfo {
	var evt MouseEventInfo
	evt.Action = uint32(wParam)
	evt.Point = info.pt
	_, high := Words(info.mouseData)
	evt.WheelDelta = high
	evt.Time = info.time

	return evt
}

func isValidEvent(m MouseAction) bool {
	return m == MOUSEMOVE ||
		m == LBUTTONDOWN ||
		m == LBUTTONUP ||
		m == RBUTTONDOWN ||
		m == RBUTTONUP ||
		m == VMOUSEWHEEL ||
		m == HMOUSEWHEEL
}

func makeHandler(cb func(MouseEventInfo) bool) wintypes.HOOKLLPROC {
	return func(code int, wParam wintypes.WPARAM, lParam unsafe.Pointer) wintypes.LRESULT {
		if code == 0 && lParam != nil && isValidEvent(uint32(wParam)) {
			evt := *(*_MSLLHOOKSTRUCT)(lParam)
			eventInfo := WhMouseEvent(wParam, evt)
			if cb(eventInfo) {
				return 1
			}
		}

		return winapi.CallNextHookEx(0, int(code), wParam, wintypes.LPARAM(lParam))
	}
}

type MouseHook struct {
	hook uintptr
}

func InstallHook(cb func(MouseEventInfo) bool) (hook *MouseHook, err error) {
	var hk MouseHook
	handler := makeHandler(cb)
	callback := syscall.NewCallback(handler)
	hk.hook = uintptr(winapi.SetWindowsHookExW(wintypes.WH_MOUSE_LL, callback, 0, 0))
	if hk.hook == 0 {
		err = fmt.Errorf("Failed to install hook.")
		return
	}

	input.KeepMessageQueuesFlushed(1)

	hook = &hk
	return
}

func (k *MouseHook) Uninstall() error {
	winapi.UnhookWindowsHookEx(wintypes.HHOOK(k.hook))
	k.hook = 0
	input.KeepMessageQueuesFlushed(-1)
	return nil
}
