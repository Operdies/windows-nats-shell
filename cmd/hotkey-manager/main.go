package main

import (
	"github.com/operdies/windows-nats-shell/cmd/hotkey-manager/border-control"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
	"github.com/operdies/windows-nats-shell/pkg/winapi"
)

// Virtual keycodes
type VKEY = int

const (
	BACKSPACE VKEY = 8
	WIN            = 91
	A              = 65
	B              = 66
	SHIFT          = 16
	ALT            = 18
	CTRL           = 17
)

func main() {
	mods := map[VKEY]bool{
		WIN:   false,
		SHIFT: false,
		ALT:   false,
		CTRL:  false,
	}

	c := client.Default()
	c.Subscribe.WH_KEYBOARD(func(kei shell.KeyboardEventInfo) {
		// combo: alt + backspace

		// Initial click
		isPress := kei.PreviousKeyState == false
		isRelease := kei.PreviousKeyState && kei.TransitionState
		if isPress || isRelease {
			k := VKEY(kei.VirtualKeyCode)
			if _, ok := mods[k]; ok {
				mods[k] = isPress
			} else if k == B {
				if mods[WIN] {
					foc := winapi.GetForegroundWindow()
					if mods[SHIFT] {
						border.Disable(foc)
					} else {
						border.Enable(foc)
					}
				}
			}
		}
	})
	select {}
}
