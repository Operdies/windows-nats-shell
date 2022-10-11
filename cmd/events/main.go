package main

import (
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
)

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

var (
	cl client.Client
)

func init() {
	cl = client.Default()
}

type eventType = int

const (
	WH_SHELL    = 1
	WH_KEYBOARD = 2
)

func publishEvent(evt eventType, nCode, wParam, lParam int) {
	if evt == WH_SHELL {
		cl.Publish.WH_SHELL(shell.WhShellEvent(int(nCode), uintptr(wParam), uintptr(lParam)))
	} else if evt == WH_KEYBOARD {
		cl.Publish.WH_KEYBOARD(shell.WhKeyboardEvent(int(nCode), uintptr(wParam), uintptr(lParam)))
	}
}

func listen() {
	// Let defers run their course when a signal is received
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func registerAsDefautlShell() {
	var min tagMINIMIZEDMETRICS
	min.iArrange = ARW_HIDE
	min.cbSize = uint32(unsafe.Sizeof(min))

	minptr := unsafe.Pointer(&min)

	systemParametersInfoA.Call(SPI_SETMINIMIZEDMETRICS, 0, uintptr(minptr), 0)
}

func main() {
	listen()
}
