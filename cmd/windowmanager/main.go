package main

import (
	"github.com/operdies/windows-nats-shell/cmd/windowmanager/inputhandler"
	"github.com/operdies/windows-nats-shell/pkg/input/keyboard"
	"github.com/operdies/windows-nats-shell/pkg/input/mouse"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
	"github.com/operdies/windows-nats-shell/pkg/winapi/windowmanager"
	"github.com/operdies/windows-nats-shell/pkg/winapi/wintypes"
)

func main() {
	nc := client.Default()
	sub, err := nc.Subscribe.WH_SHELL(func(sei shell.ShellEventInfo) {
		if sei.ShellCode == shell.HSHELL_WINDOWCREATED {
			windowmanager.HideBorder(wintypes.HWND(sei.WParam))
		}
	})
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	inputHandler := inputhandler.CreateInputHandler()
	mouseHook, _ := mouse.InstallHook(inputHandler.OnMouseInput)
	defer mouseHook.Uninstall()
	keyHook, _ := keyboard.InstallHook(inputHandler.OnKeyboardInput)
	defer keyHook.Uninstall()

	select {}
}
