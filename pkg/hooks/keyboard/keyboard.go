package keyboard

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/operdies/windows-nats-shell/pkg/winapi"
	"github.com/operdies/windows-nats-shell/pkg/winapi/wintypes"
)

type KeyboardEventInfo struct {
	ScanCode uint64
	// The scan code. The value depends on the OEM.
	VirtualKeyCode uint64
	// Indicates whether the key is an extended key, such as a function key or a key on the numeric keypad. The value is 1 if the key is an extended key; otherwise, it is 0.
	IsExtended bool
	// True if ALT is down, otherwise 0
	ContextCode bool
	// The transition state. The value is 0 if the key is being pressed and 1 if it is being released.
	TransitionState bool
	// The time stamp for this message
	Time wintypes.DWORD
}

type _KBDLLHOOKSTRUCT struct {
	VkCode      wintypes.DWORD
	ScanCode    wintypes.DWORD
	Flags       wintypes.DWORD
	Time        wintypes.DWORD
	DwExtraInfo uintptr
}

func bitRange(number uint64, start, end uint8) uint64 {
	var mask uint64
	var n uint64
	n = number >> start
	rng := end - start
	mask = (1 << (rng + 1)) - 1
	return n & mask
}

func WhKeyboardLlEvent(nCode int, info _KBDLLHOOKSTRUCT) KeyboardEventInfo {
	var evt KeyboardEventInfo
	evt.ScanCode = uint64(info.ScanCode)
	evt.VirtualKeyCode = uint64(info.VkCode)
	evt.IsExtended = bitRange(uint64(info.Flags), 0, 0) == 1
	evt.ContextCode = bitRange(uint64(info.Flags), 5, 5) == 1
	evt.TransitionState = bitRange(uint64(info.Flags), 7, 7) == 1
	evt.Time = info.Time
	return evt
}

func makeHandler(cb func(KeyboardEventInfo) bool) wintypes.HOOKLLPROC {
	return func(code int, wParam wintypes.WPARAM, lParam unsafe.Pointer) wintypes.LRESULT {
		if code == 0 && lParam != nil {
			evt := *(*_KBDLLHOOKSTRUCT)(lParam)
			eventInfo := WhKeyboardLlEvent(int(code), evt)
			if cb(eventInfo) {
				return 1
			}
		}

		return winapi.CallNextHookEx(0, int(code), wParam, wintypes.LPARAM(lParam))
	}
}

type KeyboardHook struct {
	hook uintptr
}

func InstallHook(cb func(KeyboardEventInfo) bool) (hook *KeyboardHook, err error) {
	var hk KeyboardHook
	handler := makeHandler(cb)
	callback := syscall.NewCallback(handler)
	hk.hook = uintptr(winapi.SetWindowsHookExW(wintypes.WH_KEYBOARD_LL, callback, 0, 0))
	if hk.hook == 0 {
		err = fmt.Errorf("Failed to install hook.")
		return
	}

	go func() {
		// Indefinitely process events
		// Otherwise, KeyboardEventsLl won't fire
		var msg *wintypes.MSG
		for hk.hook != 0 {
			result := winapi.GetMessage(&msg, 0, 0, 0)
			// Ignore any errors
			if result > 0 {
				winapi.TranslateMessage(&msg)
				winapi.DispatchMessageW(&msg)
			}
		}
	}()

	hook = &hk
	return
}

func (k *KeyboardHook) Uninstall() error {
	winapi.UnhookWindowsHookEx(wintypes.HHOOK(k.hook))
	k.hook = 0
	return nil
}
