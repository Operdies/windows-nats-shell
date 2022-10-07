package client

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/windows"
	"github.com/operdies/windows-nats-shell/pkg/utils"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"
)

func (client Requester) Windows() []wintypes.Window {
	response, _ := client.nc.Request(windows.GetWindows, nil, client.timeout)
	return utils.DecodeAny[[]wintypes.Window](response.Data)
}

func (client Subscriber) WindowsUpdated(callback func([]wintypes.Window)) (*nats.Subscription, error) {
	return client.nc.Subscribe(windows.WindowsUpdated, func(m *nats.Msg) {
		windows := utils.DecodeAny[[]wintypes.Window](m.Data)
		callback(windows)
	})
}

func (client Requester) SetFocus(window uint64) bool {
	msg, _ := client.nc.Request(windows.SetFocus, utils.EncodeAny(window), time.Second)
	return utils.DecodeAny[bool](msg.Data)
}

func (client Subscriber) GetWindows(callback func() []wintypes.Window) (*nats.Subscription, error) {
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

func (client Publisher) WindowsUpdated(w []wintypes.Window) {
	client.nc.Publish(windows.WindowsUpdated, utils.EncodeAny(w))
}
