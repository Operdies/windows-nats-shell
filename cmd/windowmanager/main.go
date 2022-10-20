package main

import (
	"github.com/operdies/windows-nats-shell/cmd/windowmanager/inputhandler"
	"github.com/operdies/windows-nats-shell/cmd/windowmanager/windowmanager"
	"github.com/operdies/windows-nats-shell/pkg/input/keyboard"
	"github.com/operdies/windows-nats-shell/pkg/input/mouse"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
)

func main() {
	nc := client.Default()
	cfg := client.GetConfig[windowmanager.Config](nc.Request)
	wm := windowmanager.Create(cfg)
	wm.Monitor()
	defer wm.Close()

	inputHandler := inputhandler.Create(wm)
	mouseHook, _ := mouse.InstallHook(inputHandler.OnMouseInput)
	defer mouseHook.Uninstall()
	keyHook, _ := keyboard.InstallHook(inputHandler.OnKeyboardInput)
	defer keyHook.Uninstall()

	select {}
}
