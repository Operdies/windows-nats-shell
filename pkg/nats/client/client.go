package client

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/system"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/windows"
	"github.com/operdies/windows-nats-shell/pkg/utils"
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

func (client Subscriber) WindowsUpdated(callback func(Windows)) (*nats.Subscription, error) {
	return client.nc.Subscribe(windows.WindowsUpdated, func(m *nats.Msg) {
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

func (client Requester) SetFocus(window uint64) bool {
	msg, _ := client.nc.Request(windows.SetFocus, utils.EncodeAny(window), time.Second)
	return utils.DecodeAny[bool](msg.Data)
}

func (client Client) Close() {
	client.nc.Close()
}

func New(url string, timeout time.Duration) (c Client, err error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return
	}
	c.Publish = &Publisher{nc, timeout}
	c.Request = &Requester{nc, timeout}
	c.Subscribe = &Subscriber{nc, timeout}
	return
}

func (client Subscriber) GetWindows(callback func() Windows) (*nats.Subscription, error) {
	return client.nc.Subscribe(windows.GetWindows, func(m *nats.Msg) {
		windows := callback()
		m.Respond(utils.EncodeAny(windows))
	})
}

func (client Subscriber) IsWindowFocused(callback func(wintypes.HWND) bool) (*nats.Subscription, error) {
	return client.nc.Subscribe(windows.IsWindowFocused, func(m *nats.Msg) {
		window := utils.DecodeAny[wintypes.HWND](m.Data)
		result := callback(window)
		m.Respond(utils.EncodeAny(result))
	})
}

func (client Subscriber) SetFocus(callback func(wintypes.HWND) bool) (*nats.Subscription, error) {
	return client.nc.Subscribe(windows.SetFocus, func(m *nats.Msg) {
		window := utils.DecodeAny[wintypes.HWND](m.Data)
		result := callback(window)
		m.Respond(utils.EncodeAny(result))
	})
}

func (client Subscriber) GetPrograms(callback func() []string) (*nats.Subscription, error) {
	return client.nc.Subscribe(system.GetPrograms, func(m *nats.Msg) {
		windows := callback()
		m.Respond(utils.EncodeAny(windows))
	})
}

func (client Subscriber) LaunchProgram(callback func(string) string) (*nats.Subscription, error) {
	return client.nc.Subscribe(system.LaunchProgram, func(msg *nats.Msg) {
		program := utils.DecodeAny[string](msg.Data)
		response := callback(program)
		msg.Respond(utils.EncodeAny(response))
	})
}

func (client Publisher) WindowsUpdated(w Windows) {
	client.nc.Publish(windows.WindowsUpdated, utils.EncodeAny(w))
}

// Shell
func (client Subscriber) RestartService(callback func(string) error) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.RestartService, func(msg *nats.Msg) {
		err := callback(utils.DecodeAny[string](msg.Data))
		msg.Respond(response(err))
	})
}
func (client Publisher) RestartService(service string) {
	client.nc.Publish(shell.RestartService, utils.EncodeAny(service))
}

func (client Subscriber) StopService(callback func(string) error) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.StopService, func(msg *nats.Msg) {
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

func (client Subscriber) StartService(callback func(string) error) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.StartService, func(msg *nats.Msg) {
		err := callback(utils.DecodeAny[string](msg.Data))
		msg.Respond(response(err))
	})
}

func (client Publisher) StartService(service string) {
	client.nc.Publish(shell.StartService, utils.EncodeAny(service))
}

func (client Subscriber) RestartShell(callback func() error) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.RestartShell, func(msg *nats.Msg) {
		err := callback()
		msg.Respond(response(err))
	})
}

func (client Publisher) RestartShell() {
	client.nc.Publish(shell.RestartShell, []byte{})
}

func (client Subscriber) QuitShell(callback func() error) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.QuitShell, func(msg *nats.Msg) {
		err := callback()
		msg.Respond(response(err))
	})
}

func (client Requester) QuitShell() error {
	msg, _ := client.nc.Request(shell.QuitShell, nil, client.timeout)
	return utils.DecodeAny[error](msg.Data)
}

func (client Subscriber) Config(callback func(string) *shell.Service) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.GetService, func(msg *nats.Msg) {
		key := utils.DecodeAny[string](msg.Data)
		config := callback(key)
		msg.Respond(utils.EncodeAny(config))
	})
}

// Get the request for the named service.
// If the empty string is specified, this function attempts to
// find the currently executing service based on the environment
// variable
func (client Requester) Config(name string) (service shell.Service, err error) {
	if name == "" {
		name = os.Getenv(shell.SERVICE_ENV_KEY)
	}
	if name == "" {
		fmt.Println("No such config: ", name)
		return
	}
	msg, err := client.nc.Request(shell.GetService, []byte(name), client.timeout)
	if err != nil {
		return
	}
	service = utils.DecodeAny[shell.Service](msg.Data)
	return
}

func (client Subscriber) ShellConfig(callback func() shell.Configuration) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.ShellConfig, func(msg *nats.Msg) {
		config := callback()
		msg.Respond(utils.EncodeAny(config))
	})
}

func (client Publisher) WH_SHELL(evt shell.ShellEventInfo) {
	client.nc.Publish(shell.ShellEvent, utils.EncodeAny(evt))
}

func (client Subscriber) WH_SHELL(callback func(shell.ShellEventInfo)) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.ShellEvent, func(msg *nats.Msg) {
		evt := utils.DecodeAny[shell.ShellEventInfo](msg.Data)
		callback(evt)
	})
}

func (client Publisher) WH_CBT(evt shell.CBTEventInfo) {
	client.nc.Publish(shell.CBTEvent, utils.EncodeAny(evt))
}

func (client Subscriber) WH_CBT(callback func(shell.CBTEventInfo)) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.CBTEvent, func(msg *nats.Msg) {
		evt := utils.DecodeAny[shell.CBTEventInfo](msg.Data)
		callback(evt)
	})
}

func (client Publisher) WH_KEYBOARD(evt shell.KeyboardEventInfo) {
	client.nc.Publish(shell.KeyboardEvent, utils.EncodeAny(evt))
}

func (client Subscriber) WH_KEYBOARD(callback func(shell.KeyboardEventInfo)) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.KeyboardEvent, func(msg *nats.Msg) {
		evt := utils.DecodeAny[shell.KeyboardEventInfo](msg.Data)
		callback(evt)
	})
}
