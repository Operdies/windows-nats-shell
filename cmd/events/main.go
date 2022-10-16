package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/natefinch/npipe"
	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/pkg/keyboard"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
	"github.com/operdies/windows-nats-shell/pkg/utils/query"
	"github.com/operdies/windows-nats-shell/pkg/winapi"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"
)

var (
	user32                = syscall.MustLoadDLL("user32.dll")
	systemParametersInfoA = user32.MustFindProc("SystemParametersInfoA")

	hookDll   = syscall.MustLoadDLL("libhook")
	shellProc = hookDll.MustFindProc("ShellProc")
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

type config struct {
	KeyboardEventsLL bool
	ShellEvents      bool
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

var nc client.Client

func init() {
	nc, _ = client.New(nats.DefaultURL, time.Second)
}

func publishEvent(eventType string, arguments []string) {
	numbers := query.Select(arguments, func(n string) uint64 {
		r, _ := strconv.ParseUint(n, 10, 64)
		return r
	})
	nCode, wParam, lParam := numbers[0], numbers[1], numbers[2]

	if eventType == "WH_SHELL" {
		nc.Publish.WH_SHELL(shell.WhShellEvent(int(nCode), uintptr(wParam), uintptr(lParam)))
	}
}

func handleConn(conn net.Conn, id int) {
	defer conn.Close()
	s := bufio.NewScanner(conn)
	for s.Scan() {
		msg := strings.Split(s.Text(), ",")
		publishEvent(msg[0], msg[1:])
	}
}

func connectionListener(ln *npipe.PipeListener, id int) {
	for {
		conn, err := ln.Accept()
		go func() {
			if err != nil {
				fmt.Printf("err: %v\n", err.Error())
			} else {
				handleConn(conn, id)
			}
		}()
	}
}

func server() {
	ln, err := npipe.Listen(`\\.\pipe\shellpipe`)
	if err != nil {
		panic(err)
	}

	// Start several concurrent connection listeners
	// This is to counter the case when a new connection is established
	// before the loop rolls around to accept new connections
	for i := 0; i < 5; i++ {
		go connectionListener(ln, i)
	}
}

func main() {
	cfg := client.GetConfig[config](nc.Request)
	registerShell()

	if cfg.ShellEvents {
		hook := winapi.SetWindowsHookExW(wintypes.WH_SHELL, shellProc.Addr(), wintypes.HINSTANCE(hookDll.Handle), 0)
		defer winapi.UnhookWindowsHookEx(hook)
		go server()
	}

	if cfg.KeyboardEventsLL {
		hook, _ := keyboard.InstallHook(func(kei shell.KeyboardEventInfo) bool {
			nc.Publish.WH_KEYBOARD(kei)
			return false
		})
		defer keyboard.UninstallHook(hook)
	}

	select {}
}
