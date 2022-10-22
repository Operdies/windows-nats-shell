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
	msg, _ := client.nc.Request(windows.FocusWindow, utils.EncodeAny(window), time.Second)
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
	return client.nc.Subscribe(windows.FocusWindow, func(m *nats.Msg) {
		window := utils.DecodeAny[wintypes.HWND](m.Data)
		result := callback(window)
		m.Respond(utils.EncodeAny(result))
	})
}

func (client Publisher) WindowsUpdated(w []wintypes.Window) {
	client.nc.Publish(windows.WindowsUpdated, utils.EncodeAny(w))
}

func (client Subscriber) HideBorder(callback func(wintypes.HWND) bool) (*nats.Subscription, error) {
	return client.nc.Subscribe(windows.HideBorder, func(msg *nats.Msg) {
		result := callback(utils.DecodeAny[wintypes.HWND](msg.Data))
		msg.Respond(utils.EncodeAny(result))
	})
}

func (client Requester) HideBorder(hwnd wintypes.HWND) bool {
	msg, _ := client.nc.Request(windows.HideBorder, utils.EncodeAny(hwnd), client.timeout)
	return utils.DecodeAny[bool](msg.Data)
}

func (client Subscriber) ShowBorder(callback func(wintypes.HWND) bool) (*nats.Subscription, error) {
	return client.nc.Subscribe(windows.ShowBorder, func(msg *nats.Msg) {
		result := callback(utils.DecodeAny[wintypes.HWND](msg.Data))
		msg.Respond(utils.EncodeAny(result))
	})
}

func (client Requester) ShowBorder(hwnd wintypes.HWND) bool {
	msg, _ := client.nc.Request(windows.ShowBorder, utils.EncodeAny(hwnd), client.timeout)
	return utils.DecodeAny[bool](msg.Data)
}

func (client Subscriber) ToggleBorder(callback func(wintypes.HWND) bool) (*nats.Subscription, error) {
	return client.nc.Subscribe(windows.ToggleBorder, func(msg *nats.Msg) {
		result := callback(utils.DecodeAny[wintypes.HWND](msg.Data))
		msg.Respond(utils.EncodeAny(result))
	})
}

func (client Requester) ToggleBorder(hwnd wintypes.HWND) bool {
	msg, _ := client.nc.Request(windows.ToggleBorder, utils.EncodeAny(hwnd), client.timeout)
	return utils.DecodeAny[bool](msg.Data)
}

// func (client Subscriber) MoveWindow(callback func(windows.MoveEvent) bool) (*nats.Subscription, error) {
// 	return client.nc.Subscribe(windows.MoveWindow, func(msg *nats.Msg) {
// 		result := callback(utils.DecodeAny[windows.MoveEvent](msg.Data))
// 		msg.Respond(utils.EncodeAny(result))
// 	})
// }
//
// func (client Requester) MoveWindow(me windows.MoveEvent) bool {
// 	msg, _ := client.nc.Request(windows.MoveWindow, utils.EncodeAny(me), client.timeout)
// 	return utils.DecodeAny[bool](msg.Data)
// }
//
// func (client Subscriber) ResizeWindow(callback func(windows.ResizeEvent) bool) (*nats.Subscription, error) {
// 	return client.nc.Subscribe(windows.ResizeWindow, func(msg *nats.Msg) {
// 		result := callback(utils.DecodeAny[windows.ResizeEvent](msg.Data))
// 		msg.Respond(utils.EncodeAny(result))
// 	})
// }
//
// func (client Requester) ResizeWindow(resizeEvent windows.ResizeEvent) bool {
// 	msg, _ := client.nc.Request(windows.ResizeWindow, utils.EncodeAny(resizeEvent), client.timeout)
// 	return utils.DecodeAny[bool](msg.Data)
// }

// func (client Subscriber) MinimizeWindow(callback func(wintypes.HWND) bool) (*nats.Subscription, error) {
// 	return client.nc.Subscribe(windows.MinimizeWindow, func(msg *nats.Msg) {
// 		result := callback(utils.DecodeAny[wintypes.HWND](msg.Data))
// 		msg.Respond(utils.EncodeAny(result))
// 	})
// }
//
// func (client Requester) MinimizeWindow(hwnd wintypes.HWND) bool {
// 	msg, _ := client.nc.Request(windows.MinimizeWindow, utils.EncodeAny(hwnd), client.timeout)
// 	return utils.DecodeAny[bool](msg.Data)
// }
//
// func (client Subscriber) RestoreWindow(callback func(wintypes.HWND) bool) (*nats.Subscription, error) {
// 	return client.nc.Subscribe(windows.RestoreWindow, func(msg *nats.Msg) {
// 		result := callback(utils.DecodeAny[wintypes.HWND](msg.Data))
// 		msg.Respond(utils.EncodeAny(result))
// 	})
// }
//
// func (client Requester) RestoreWindow(hwnd wintypes.HWND) bool {
// 	msg, _ := client.nc.Request(windows.RestoreWindow, utils.EncodeAny(hwnd), client.timeout)
// 	return utils.DecodeAny[bool](msg.Data)
// }
//
// func (client Subscriber) MaximizeWindow(callback func(wintypes.HWND) bool) (*nats.Subscription, error) {
// 	return client.nc.Subscribe(windows.MaximizeWindow, func(msg *nats.Msg) {
// 		result := callback(utils.DecodeAny[wintypes.HWND](msg.Data))
// 		msg.Respond(utils.EncodeAny(result))
// 	})
// }
//
// func (client Requester) MaximizeWindow(hwnd wintypes.HWND) bool {
// 	msg, _ := client.nc.Request(windows.MaximizeWindow, utils.EncodeAny(hwnd), client.timeout)
// 	return utils.DecodeAny[bool](msg.Data)
// }
//
// func (client Subscriber) FocusPrevious(callback func(wintypes.HWND) bool) (*nats.Subscription, error) {
// 	return client.nc.Subscribe(windows.FocusPrevious, func(msg *nats.Msg) {
// 		result := callback(utils.DecodeAny[wintypes.HWND](msg.Data))
// 		msg.Respond(utils.EncodeAny(result))
// 	})
// }
//
// func (client Requester) FocusPrevious(hwnd wintypes.HWND) bool {
// 	msg, _ := client.nc.Request(windows.FocusPrevious, utils.EncodeAny(hwnd), client.timeout)
// 	return utils.DecodeAny[bool](msg.Data)
// }
//
// func (client Subscriber) FocusNext(callback func(wintypes.HWND) bool) (*nats.Subscription, error) {
// 	return client.nc.Subscribe(windows.FocusNext, func(msg *nats.Msg) {
// 		result := callback(utils.DecodeAny[wintypes.HWND](msg.Data))
// 		msg.Respond(utils.EncodeAny(result))
// 	})
// }
//
// func (client Requester) FocusNext(hwnd wintypes.HWND) bool {
// 	msg, _ := client.nc.Request(windows.FocusNext, utils.EncodeAny(hwnd), client.timeout)
// 	return utils.DecodeAny[bool](msg.Data)
// }
//
