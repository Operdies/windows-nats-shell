package input

import (
	"sync"

	"github.com/operdies/windows-nats-shell/pkg/winapi"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"
)

type VKEY uint32

var (
	VK_MAP = map[string]VKEY{
		"nullkey":   255,
		"backspace": VK_BACK,
		"tab":       VK_TAB,
		"return":    VK_RETURN,
		"pause":     VK_PAUSE,
		"enter":     VK_RETURN,
		"escape":    VK_ESCAPE,
		"space":     VK_SPACE,

		"pgdn":       VK_PRIOR,
		"pagedown":   VK_PRIOR,
		"pageup":     VK_NEXT,
		"pgup":       VK_NEXT,
		"end":        VK_END,
		"home":       VK_HOME,
		"numlock":    VK_NUMLOCK,
		"scrolllock": VK_SCROLL,

		"left":  VK_LEFT,
		"up":    VK_UP,
		"right": VK_RIGHT,
		"down":  VK_DOWN,

		"print":       VK_SNAPSHOT,
		"printscreen": VK_SNAPSHOT,
		"insert":      VK_INSERT,
		"del":         VK_DELETE,
		"delete":      VK_DELETE,

		"shift":    VK_LSHIFT,
		"lshift":   VK_LSHIFT,
		"rshift":   VK_RSHIFT,
		"lctrl":    VK_LCONTROL,
		"ctrl":     VK_LCONTROL,
		"control":  VK_LCONTROL,
		"lcontrol": VK_LCONTROL,
		"rctrl":    VK_RCONTROL,
		"rcontrol": VK_RCONTROL,
		"alt":      VK_LMENU,
		"menu":     VK_LMENU,
		"win":      VK_LWIN,
		"lwin":     VK_LWIN,
		"rwin":     VK_RWIN,

		"num0": VK_NUMPAD0,
		"num1": VK_NUMPAD1,
		"num2": VK_NUMPAD2,
		"num3": VK_NUMPAD3,
		"num4": VK_NUMPAD4,
		"num5": VK_NUMPAD5,
		"num6": VK_NUMPAD6,
		"num7": VK_NUMPAD7,
		"num8": VK_NUMPAD8,
		"num9": VK_NUMPAD9,

		"f1":  VK_F1,
		"f2":  VK_F2,
		"f3":  VK_F3,
		"f4":  VK_F4,
		"f5":  VK_F5,
		"f6":  VK_F6,
		"f7":  VK_F7,
		"f8":  VK_F8,
		"f9":  VK_F9,
		"f10": VK_F10,
		"f11": VK_F11,
		"f12": VK_F12,
	}
)

const (
	VK_BACK   VKEY = 0x08 //backspace
	VK_TAB    VKEY = 0x09
	VK_RETURN VKEY = 0x0D
	VK_PAUSE  VKEY = 0x13
	VK_ESCAPE VKEY = 0x1B
	VK_SPACE  VKEY = 0x20

	VK_PRIOR   VKEY = 0x21 //pageup
	VK_NEXT    VKEY = 0x22 //pagedown
	VK_END     VKEY = 0x23
	VK_HOME    VKEY = 0x24
	VK_NUMLOCK VKEY = 0x90
	VK_SCROLL  VKEY = 0x91 // scroll lock

	VK_LEFT  VKEY = 0x25
	VK_UP    VKEY = 0x26
	VK_RIGHT VKEY = 0x27
	VK_DOWN  VKEY = 0x28

	VK_SNAPSHOT VKEY = 0x2C // print screen
	VK_INSERT   VKEY = 0x2D
	VK_DELETE   VKEY = 0x2E

	VK_SHIFT   VKEY = 0x10
	VK_CONTROL VKEY = 0x11
	VK_MENU    VKEY = 0x12 // alt
	VK_LWIN    VKEY = 0x5B
	VK_RWIN    VKEY = 0x5C

	/* Numpad keys */
	VK_NUMPAD0   VKEY = 0x60
	VK_NUMPAD1   VKEY = 0x61
	VK_NUMPAD2   VKEY = 0x62
	VK_NUMPAD3   VKEY = 0x63
	VK_NUMPAD4   VKEY = 0x64
	VK_NUMPAD5   VKEY = 0x65
	VK_NUMPAD6   VKEY = 0x66
	VK_NUMPAD7   VKEY = 0x67
	VK_NUMPAD8   VKEY = 0x68
	VK_NUMPAD9   VKEY = 0x69
	VK_MULTIPLY  VKEY = 0x6a
	VK_ADD       VKEY = 0x6b
	VK_SEPARATOR VKEY = 0x6c
	VK_SUBTRACT  VKEY = 0x6d
	VK_DECIMAL   VKEY = 0x6e
	VK_DIVIDE    VKEY = 0x6f

	VK_F1  VKEY = 0x70
	VK_F2  VKEY = 0x71
	VK_F3  VKEY = 0x72
	VK_F4  VKEY = 0x73
	VK_F5  VKEY = 0x74
	VK_F6  VKEY = 0x75
	VK_F7  VKEY = 0x76
	VK_F8  VKEY = 0x77
	VK_F9  VKEY = 0x78
	VK_F10 VKEY = 0x79
	VK_F11 VKEY = 0x7a
	VK_F12 VKEY = 0x7b

	VK_LSHIFT   VKEY = 0xA0
	VK_RSHIFT   VKEY = 0xA1
	VK_LCONTROL VKEY = 0xA2
	VK_RCONTROL VKEY = 0xA3
	VK_LMENU    VKEY = 0xA4
	VK_RMENU    VKEY = 0xA5
)

var (
	flushCounter = 0
	mut          = sync.Mutex{}
)

func flushWhileNeeded() {
	var msg *wintypes.MSG
	for flushCounter > 0 {
		result := winapi.GetMessage(&msg, 0, 0, 0)
		// Ignore any errors
		if result > 0 {
			winapi.TranslateMessage(&msg)
			winapi.DispatchMessageW(&msg)
		}
	}
}

func KeepMessageQueuesFlushed(n int) {
	mut.Lock()
	defer mut.Unlock()
	flushCounter += n
	if flushCounter > 0 {
		go flushWhileNeeded()
	}
}
