package main

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
	"github.com/operdies/windows-nats-shell/cmd/shell-event-publisher/hooks"
	"gopkg.in/natefinch/npipe.v2"
)

var (
	cl client.Client
)

func init() {
	cl, _ = client.New(nats.DefaultURL, time.Second)
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	s := bufio.NewScanner(conn)
	for s.Scan() {
		cl.Publish.ShellEvent(shell.Event{Event: s.Text()})
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

func main() {
  hooks.Register()
  defer hooks.Unregister()
	server()
}
