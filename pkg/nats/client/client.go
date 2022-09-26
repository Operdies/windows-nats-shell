package client

import (
	"errors"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/pkg/nats/api"
	"github.com/operdies/windows-nats-shell/pkg/nats/utils"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"
)

const (
	timeout = time.Second
)

type Client struct {
	nc *nats.Conn
}

type Windows = []wintypes.Window

func (client Client) Windows() Windows {
	response, _ := client.nc.Request(api.GetWindows, nil, timeout)
	return utils.DecodeAny[Windows](response.Data)
}

func (client Client) OnWindowsUpdated(callback func(Windows)) {
	client.nc.Subscribe(api.WindowsUpdated, func(m *nats.Msg) {
		windows := utils.DecodeAny[Windows](m.Data)
		callback(windows)
	})
}

func (client Client) GetPrograms() []string {
	msg, _ := client.nc.Request(api.GetPrograms, nil, timeout*5)
	programs := utils.DecodeAny[[]string](msg.Data)
	return programs
}

func (client Client) LaunchProgram(program string) error {
	msg, _ := client.nc.Request(api.LaunchProgram, utils.EncodeAny(program), timeout*2)
	status := utils.DecodeAny[string](msg.Data)
	if status == "Ok" {
		return nil
	}
	return errors.New(string(msg.Data))
}

func (client Client) SetFocus(window uint64) {
	client.nc.Request(api.SetFocus, utils.EncodeAny(window), time.Second)
}

func (client Client) Close() {
	client.nc.Close()
}

func New(url string) (c Client, err error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return
	}
	c.nc = nc
	return
}

func (client Client) OnGetWindows(callback func() Windows) {
	client.nc.Subscribe(api.GetWindows, func(m *nats.Msg) {
		windows := callback()
		m.Respond(utils.EncodeAny(windows))
	})
}

func (client Client) OnIsWindowFocused(callback func(wintypes.HWND) bool) {
	client.nc.Subscribe(api.IsWindowFocused, func(m *nats.Msg) {
		window := utils.DecodeAny[wintypes.HWND](m.Data)
		result := callback(window)
		m.Respond(utils.EncodeAny(result))
	})
}

func (client Client) OnSetFocus(callback func(wintypes.HWND) bool) {
	client.nc.Subscribe(api.SetFocus, func(m *nats.Msg) {
		window := utils.DecodeAny[wintypes.HWND](m.Data)
		result := callback(window)
		m.Respond(utils.EncodeAny(result))
	})
}

func (client Client) OnGetPrograms(callback func() []string) {
	client.nc.Subscribe(api.GetPrograms, func(m *nats.Msg) {
		windows := callback()
		m.Respond(utils.EncodeAny(windows))
	})
}

func (client Client) OnLaunchProgram(callback func(string) string) {
	client.nc.Subscribe(api.LaunchProgram, func(msg *nats.Msg) {
		program := utils.DecodeAny[string](msg.Data)
		response := callback(program)
		msg.Respond(utils.EncodeAny(response))
	})
}

func (client Client) PublishWindows(w Windows) {
	client.nc.Publish(api.GetWindows, utils.EncodeAny(w))
}
