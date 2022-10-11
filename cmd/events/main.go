package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
	"github.com/operdies/windows-nats-shell/pkg/utils/query"
	"github.com/operdies/windows-nats-shell/pkg/winapi"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"

	"gopkg.in/natefinch/npipe.v2"
)

const (
	ShellProc = "ShellProc"
	KeyboardProc = "KeyboardProc"
)

var (
	hookDll = syscall.MustLoadDLL("libhook")

	shellProc = hookDll.MustFindProc(ShellProc)
	keyboardProc = hookDll.MustFindProc(KeyboardProc)
)

var (
	cl client.Client
)

func init() {
	cl, _ = client.New(nats.DefaultURL, time.Second)
}

func publishEvent(eventType string, arguments []string) {
	numbers := query.Select(arguments, func(n string) uint64 {
		r, _ := strconv.ParseUint(n, 10, 64)
		return r
	})
	nCode, wParam, lParam := numbers[0], numbers[1], numbers[2]

	if eventType == "WH_SHELL" {
		cl.Publish.WH_SHELL(shell.WhShellEvent(int(nCode), uintptr(wParam), uintptr(lParam)))
	} else if eventType == "WH_KEYBOARD" {
		cl.Publish.WH_KEYBOARD(shell.WhKeyboardEvent(int(nCode), uintptr(wParam), uintptr(lParam)))
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
	select {}
}

func listen() {
	hook1 := winapi.SetWindowsHookExW(wintypes.WH_SHELL, shellProc.Addr(), wintypes.HINSTANCE(hookDll.Handle), 0)
	hook3 := winapi.SetWindowsHookExW(wintypes.WH_KEYBOARD, keyboardProc.Addr(), wintypes.HINSTANCE(hookDll.Handle), 0)
	go server()

	defer winapi.UnhookWindowsHook(hook1)
	// defer winapi.UnhookWindowsHook(hook2)
	defer winapi.UnhookWindowsHook(hook3)

	// Let defers run their course when a signal is received
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func main() {
	listen()
}
