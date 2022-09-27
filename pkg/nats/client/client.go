package client

import (
	"errors"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/system"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/windows"
	"github.com/operdies/windows-nats-shell/pkg/nats/utils"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"
)

var subscriber interface {
}

type Client struct {
	nc        *nats.Conn
	timeout   time.Duration
	Subscribe *Subscriber
	Publish   *Publisher
	Request   *Requester
}

type Subscriber struct {
	nc      *nats.Conn
	timeout time.Duration
}

type Publisher struct {
	nc      *nats.Conn
	timeout time.Duration
}

type Requester struct {
	nc      *nats.Conn
	timeout time.Duration
}

type Windows = []wintypes.Window

func (client Requester) Windows() Windows {
	response, _ := client.nc.Request(windows.GetWindows, nil, client.timeout)
	return utils.DecodeAny[Windows](response.Data)
}

func (client Subscriber) WindowsUpdated(callback func(Windows)) {
	client.nc.Subscribe(windows.WindowsUpdated, func(m *nats.Msg) {
		windows := utils.DecodeAny[Windows](m.Data)
		callback(windows)
	})
}

func (client Requester) GetPrograms() []string {
	msg, _ := client.nc.Request(system.GetPrograms, nil, client.timeout)
	programs := utils.DecodeAny[[]string](msg.Data)
	return programs
}

func (client Requester) LaunchProgram(program string) error {
	msg, _ := client.nc.Request(system.LaunchProgram, utils.EncodeAny(program), client.timeout)
	status := utils.DecodeAny[string](msg.Data)
	if status == "Ok" {
		return nil
	}
	return errors.New(string(msg.Data))
}

func (client Requester) SetFocus(window uint64) {
	client.nc.Request(windows.SetFocus, utils.EncodeAny(window), time.Second)
}

func (client Client) Close() {
	client.nc.Close()
}

func New(url string, timeout time.Duration) (c Client, err error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return
	}
	c.nc = nc
	c.timeout = timeout
	c.Publish = &Publisher{nc, timeout}
	c.Request = &Requester{nc, timeout}
	c.Subscribe = &Subscriber{nc, timeout}
	return
}

func (client Subscriber) GetWindows(callback func() Windows) {
	client.nc.Subscribe(windows.GetWindows, func(m *nats.Msg) {
		windows := callback()
		m.Respond(utils.EncodeAny(windows))
	})
}

func (client Subscriber) IsWindowFocused(callback func(wintypes.HWND) bool) {
	client.nc.Subscribe(windows.IsWindowFocused, func(m *nats.Msg) {
		window := utils.DecodeAny[wintypes.HWND](m.Data)
		result := callback(window)
		m.Respond(utils.EncodeAny(result))
	})
}

func (client Subscriber) SetFocus(callback func(wintypes.HWND) bool) {
	client.nc.Subscribe(windows.SetFocus, func(m *nats.Msg) {
		window := utils.DecodeAny[wintypes.HWND](m.Data)
		result := callback(window)
		m.Respond(utils.EncodeAny(result))
	})
}

func (client Subscriber) GetPrograms(callback func() []string) {
	client.nc.Subscribe(system.GetPrograms, func(m *nats.Msg) {
		windows := callback()
		m.Respond(utils.EncodeAny(windows))
	})
}

func (client Subscriber) LaunchProgram(callback func(string) string) {
	client.nc.Subscribe(system.LaunchProgram, func(msg *nats.Msg) {
		program := utils.DecodeAny[string](msg.Data)
		response := callback(program)
		msg.Respond(utils.EncodeAny(response))
	})
}

func (client Publisher) WindowsUpdated(w Windows) {
	client.nc.Publish(windows.WindowsUpdated, utils.EncodeAny(w))
}

// Shell
func (client Subscriber) RestartService(callback func(string) error) {
	client.nc.Subscribe(shell.RestartService, func(msg *nats.Msg) {
		err := callback(utils.DecodeAny[string](msg.Data))
		msg.Respond(response(err))
	})
}
func (client Publisher) RestartService(service string) {
	client.nc.Publish(shell.RestartService, utils.EncodeAny(service))
}

func (client Subscriber) StopService(callback func(string) error) {
	client.nc.Subscribe(shell.StopService, func(msg *nats.Msg) {
		err := callback(utils.DecodeAny[string](msg.Data))
		msg.Respond(response(err))
	})
}

func (client Publisher) StopService(service string) {
	client.nc.Publish(shell.StopService, utils.EncodeAny(service))
}

func response(err error) []byte {
	if err == nil {
		return []byte("Ok")
	}
	return []byte(err.Error())
}

func (client Subscriber) StartService(callback func(string) error) {
	client.nc.Subscribe(shell.StartService, func(msg *nats.Msg) {
		err := callback(utils.DecodeAny[string](msg.Data))
		msg.Respond(response(err))
	})
}

func (client Publisher) StartService(service string) {
	client.nc.Publish(shell.StartService, utils.EncodeAny(service))
}

func (client Subscriber) RestartShell(callback func() error) {
	client.nc.Subscribe(shell.RestartShell, func(msg *nats.Msg) {
		err := callback()
		msg.Respond(response(err))
	})
}

func (client Publisher) RestartShell() {
	client.nc.Publish(shell.RestartShell, []byte{})
}

func (client Subscriber) Config(callback func() shell.Configuration) {
	client.nc.Subscribe(shell.Config, func(msg *nats.Msg) {
    config := callback()
		msg.Respond(utils.EncodeAny(config))
	})
}

func (client Requester) Config() shell.Configuration {
  msg, _ := client.nc.Request(shell.Config, nil, client.timeout)
  return utils.DecodeAny[shell.Configuration](msg.Data)
}

