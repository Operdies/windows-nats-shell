package main

import (
	"fmt"
	"syscall"
)

var (
	libevents     = syscall.MustLoadDLL("libevents")
	ShellCallback = libevents.MustFindProc("PublishEvent")
)

func main() {
a, b, c := ShellCallback.Call()
	fmt.Println(a, b, c)
}
