package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/natefinch/npipe"
	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/pkg/input/keyboard"
	"github.com/operdies/windows-nats-shell/pkg/input/mouse"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
	"github.com/operdies/windows-nats-shell/pkg/utils/query"
	"github.com/operdies/windows-nats-shell/pkg/winapi"
	"github.com/operdies/windows-nats-shell/pkg/winapi/wintypes"
)

var (
	hookDll   = syscall.MustLoadDLL("libhook")
	shellProc = hookDll.MustFindProc("ShellProc")
)

type config struct {
	KeyboardEvents bool
	MouseEvents    bool
	ShellEvents    bool
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

	if cfg.ShellEvents {
		hook := winapi.SetWindowsHookExW(wintypes.WH_SHELL, shellProc.Addr(), wintypes.HINSTANCE(hookDll.Handle), 0)
		defer winapi.UnhookWindowsHookEx(hook)
		go server()
	}

	if cfg.KeyboardEvents {
		hook, _ := keyboard.InstallHook(func(kei keyboard.KeyboardEventInfo) bool {
			nc.Publish.WH_KEYBOARD(kei)
			return false
		})
		defer hook.Uninstall()
	}

	if cfg.MouseEvents {
		hook, _ := mouse.InstallHook(func(mei mouse.MouseEventInfo) bool {
			nc.Publish.WH_MOUSE(mei)
			return false
		})
		defer hook.Uninstall()
	}

	select {}
}
