package main

import (
	"github.com/operdies/windows-nats-shell/cmd/windowmanager/inputhandler"
	"github.com/operdies/windows-nats-shell/pkg/input/keyboard"
	"github.com/operdies/windows-nats-shell/pkg/input/mouse"
	// "github.com/operdies/windows-nats-shell/pkg/nats/client"
)

func main() {
	inputHandler := inputhandler.CreateInputHandler()
	mouseHook, _ := mouse.InstallHook(inputHandler.OnMouseInput)
	defer mouseHook.Uninstall()
	keyHook, _ := keyboard.InstallHook(inputHandler.OnKeyboardInput)
	defer keyHook.Uninstall()
	// nc := client.Default()
	select {}
}
