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
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
	"github.com/operdies/windows-nats-shell/pkg/utils/query"
	"github.com/operdies/windows-nats-shell/pkg/winapi"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"
)

var (
	user32                = syscall.MustLoadDLL("user32.dll")
	systemParametersInfoA = user32.MustFindProc("SystemParametersInfoA")

	hookDll      = syscall.MustLoadDLL("libhook")
	shellProc    = hookDll.MustFindProc("ShellProc")
	keyboardProc = hookDll.MustFindProc("KeyboardProc")
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
	KeyboardEvents   bool
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

func ShellHookHandler(c chan shell.ShellEventInfo) wintypes.HOOKPROC {
	return func(code int32, wParam wintypes.WPARAM, lParam wintypes.LPARAM) wintypes.LRESULT {
		if code > 0 {

		}
		return winapi.CallNextHookEx(0, int(code), wParam, lParam)
	}
}

func keyboardHandler(code int32, wParam wintypes.WPARAM, lParam wintypes.LPARAM) wintypes.LRESULT {
	if code == 0 && lParam != 0 {
		evt := *(*shell.KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		evt2 := shell.WhKeyboardLlEvent(int(code), evt)
		handled := nc.Request.WH_KEYBOARD(evt2)
		//
		if handled {
			// This doesn't actually intercept the event for other applications :(
			return wintypes.LRESULT(1)
		}
	}

	return winapi.CallNextHookEx(0, int(code), wParam, lParam)
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
	} else if eventType == "WH_KEYBOARD" {
		nc.Publish.WH_KEYBOARD(shell.WhKeyboardEvent(int(nCode), wintypes.WPARAM(wParam), wintypes.LPARAM(lParam)))
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
	cfg, _ := nc.Request.Config("")
	custom, _ := shell.GetCustom[config](cfg)

	registerShell()

	if custom.ShellEvents {
		hook := winapi.SetWindowsHookExW(wintypes.WH_SHELL, shellProc.Addr(), wintypes.HINSTANCE(hookDll.Handle), 0)
		defer winapi.UnhookWindowsHookEx(hook)
		go server()
	}

	if custom.KeyboardEventsLL {
		callback := syscall.NewCallback(keyboardHandler)
		hook := winapi.SetWindowsHookExW(wintypes.WH_KEYBOARD_LL, callback, 0, 0)
		defer winapi.UnhookWindowsHookEx(hook)
		// Indefinitely process events
		// Otherwise, KeyboardEventsLl won't fire
		var msg *wintypes.MSG
		for {
			result := winapi.GetMessage(&msg, 0, 0, 0)
			// Ignore any errors
			if result > 0 {
				winapi.TranslateMessage(&msg)
				winapi.DispatchMessageW(&msg)
			}
		}
	}
}
