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
	// "github.com/operdies/windows-nats-shell/cmd/shell-event-publisher/hooks"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
	"github.com/operdies/windows-nats-shell/pkg/winapi"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"

	// "github.com/operdies/windows-nats-shell/pkg/utils/query"
	"gopkg.in/natefinch/npipe.v2"
)

const (
	ShellProc    = "ShellProc"
	CBTProc      = "CBTProc"
	KeyboardProc = "KeyboardProc"
)

var (
	hookDll = syscall.MustLoadDLL("C:\\Users\\alexw\\repos\\minimalist-shell\\bin\\libhook.dll")

	shellProc    = hookDll.MustFindProc(ShellProc)
	cbtProc      = hookDll.MustFindProc(CBTProc)
	keyboardProc = hookDll.MustFindProc(KeyboardProc)
)

var (
	cl client.Client
)

func init() {
	cl, _ = client.New(nats.DefaultURL, time.Second)
}

// func publishEvent(eventType string, arguments []string) {
//   numbers := query.Select(arguments, func(n string) uint64 {
// 		r, _ := strconv.ParseUint(n, 10, 64)
// 		return r
// 	})
// 	if eventType == "WH_SHELL" {
// 	} else if eventType == "WH_CBT" {
// 	} else if eventType == "WH_KEYBOARD" {
// 	}
//
// }

func handleConn(conn net.Conn) {
	defer conn.Close()
	s := bufio.NewScanner(conn)
	for s.Scan() {
		msg := strings.Split(s.Text(), ",")
		parts := msg[1:]
		numbers := make([]uint64, 3)
		for i := 0; i < 3; i++ {
			numbers[i], _ = strconv.ParseUint(parts[i], 10, 64)
		}
		if msg[0] == "WH_SHELL" {
			cl.Publish.WH_SHELL(shell.WhShellEvent(int(numbers[0]), uintptr(numbers[1]), uintptr(numbers[2])))
		} else if msg[0] == "WH_CBT" {
			cl.Publish.WH_CBT(shell.WhCbtEvent(int(numbers[0]), uintptr(numbers[1]), uintptr(numbers[2])))
		} else if msg[0] == "WH_KEYBOARD" {
			cl.Publish.WH_KEYBOARD(shell.WhKeyboardEvent(int(numbers[0]), uintptr(numbers[1]), uintptr(numbers[2])))
		}

	}
}

func server() {
	ln, err := npipe.Listen(`\\.\pipe\shellpipe`)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("err: %v\n", err.Error())
			continue
		}
		go handleConn(conn)
	}
}

func listen() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	hook1 := winapi.SetWindowsHookExW(wintypes.WH_SHELL, shellProc.Addr(), wintypes.HINSTANCE(hookDll.Handle), 0)
	hook2 := winapi.SetWindowsHookExW(wintypes.WH_CBT, cbtProc.Addr(), wintypes.HINSTANCE(hookDll.Handle), 0)
	hook3 := winapi.SetWindowsHookExW(wintypes.WH_KEYBOARD, keyboardProc.Addr(), wintypes.HINSTANCE(hookDll.Handle), 0)
	go server()

	defer winapi.UnhookWindowsHook(hook1)
	defer winapi.UnhookWindowsHook(hook2)
	defer winapi.UnhookWindowsHook(hook3)

	// Let defers run their course when a signal is received
	<-c
}

func main() {
	listen()
}
