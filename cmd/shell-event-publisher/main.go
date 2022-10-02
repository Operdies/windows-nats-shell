package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/cmd/shell-event-publisher/hooks"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
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
    parts := strings.Split(s.Text(), ",")
    numbers := make([]uint64, 3)
    for i := 0; i < 3; i++ {
      numbers[i], _ = strconv.ParseUint(parts[i], 10, 64)
    }
		cl.Publish.ShellEvent(shell.NewEvent(int(numbers[0]), uintptr(numbers[1]), uintptr(numbers[2])))
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
