package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	// "github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	// "github.com/operdies/windows-nats-shell/pkg/nats/client"
	"github.com/operdies/windows-nats-shell/pkg/winapi"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"
)

import "C"

var (
	user32 = syscall.MustLoadDLL("user32.dll")

	systemParametersInfoA = user32.MustFindProc("SystemParametersInfoA")
)

type tagMINIMIZEDMETRICS struct {
	cbSize   uint32
	iWidth   int32
	iHorzGap int32
	iVertGap int32
	iArrange int32
}

const (
	ARW_HIDE                = 0x0008
	SPI_SETMINIMIZEDMETRICS = 0x002C
)

type eventType = int

const (
	WH_SHELL    = 1
	WH_KEYBOARD = 2
)

//export PublishEvent
func PublishEvent() {
	fmt.Printf("hello\n")
}
func PublishEvent2(evt eventType, nCode, wParam, lParam int) {
	fmt.Printf("hello\n")
	// cl := client.Default()
	// if evt == WH_SHELL {
	// 	cl.Publish.WH_SHELL(shell.WhShellEvent(int(nCode), uintptr(wParam), uintptr(lParam)))
	// } else if evt == WH_KEYBOARD {
	// 	cl.Publish.WH_KEYBOARD(shell.WhKeyboardEvent(int(nCode), uintptr(wParam), uintptr(lParam)))
	// }
}

//export ShellCallback
func ShellCallback(ncode int32, wParam uint64, lParam uint64) uint64 {
	if ncode >= 0 {
		PublishEvent2(WH_SHELL, int(ncode), int(wParam), int(lParam))
	}
	return uint64(winapi.CallNextHookEx(0, int(ncode), wintypes.WPARAM(wParam), wintypes.LPARAM(lParam)))
}

func listen() {
	cb := syscall.NewCallback(ShellCallback)
	shellHook := winapi.SetWindowsHookExW(WH_SHELL, cb, 0, winapi.GetCurrentThreadId())
	defer winapi.UnhookWindowsHook(shellHook)
	// Let defers run their course when a signal is received
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func registerShell() {
	var min tagMINIMIZEDMETRICS
	min.iArrange = ARW_HIDE
	min.cbSize = uint32(unsafe.Sizeof(min))

	minptr := unsafe.Pointer(&min)

	// This call is required in order to receive shell events.
	// It also hides minimized windows so there is no pseudo-taskbar
	systemParametersInfoA.Call(SPI_SETMINIMIZEDMETRICS, 0, uintptr(minptr), 0)
}

func main() {
	// listen()
}
