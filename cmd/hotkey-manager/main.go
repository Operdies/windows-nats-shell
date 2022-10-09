package main

import (
	"github.com/operdies/windows-nats-shell/cmd/hotkey-manager/keymap"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
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
	c := client.Default()
	km := keymap.Create()

	// Maybe this needs to be a WindowsHookEvent callback in the future.
	// For simplicity, let's stick to subscribing for now.
	// A windows hook event would allow us to avoid propagating handled events
	c.Subscribe.WH_KEYBOARD(func(kei shell.KeyboardEventInfo) { km.ProcessEvent(kei) })
	select {}
}
