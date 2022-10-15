package main

import (
	"fmt"

	"github.com/operdies/windows-nats-shell/cmd/hotkeys/keymap"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
)

func dumpTree(mods []uint32, bt *keymap.BindingTree) {
	for m, b := range bt.Subtrees {
		mods2 := append(mods, m)
		dumpTree(mods2, b)

		if b.HasAction {
			fmt.Printf("Binding: %+v\nActions: %+v\n", mods2, b.Action)
		}
	}
}

func main() {
	c := client.Default()
	km := keymap.Create()
	dumpTree(nil, km.Bindings)

	// Maybe this needs to be a WindowsHookEvent callback in the future.
	// For simplicity, let's stick to subscribing for now.
	// A windows hook event would allow us to avoid propagating handled events
	c.Subscribe.WH_KEYBOARD(func(kei shell.KeyboardEventInfo) bool {
		return km.ProcessEvent(kei)
	})
	select {}
}
