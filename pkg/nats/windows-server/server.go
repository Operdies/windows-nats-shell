//go:build windows && amd64
// +build windows,amd64

package server

import (
	"time"

	"github.com/nats-io/nats.go"

	"github.com/operdies/windows-nats-shell/pkg/nats/api"
	"github.com/operdies/windows-nats-shell/pkg/nats/utils"
	"github.com/operdies/windows-nats-shell/pkg/winapi"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"
)

func poll(nc *nats.Conn, interval time.Duration) {
	ticker := time.NewTicker(interval)
	prevWindows := make([]winapi.Window, 0)

	anyChanged := func(windows []winapi.Window) bool {
		if len(prevWindows) != len(windows) {
			return true
		}
		for i := 0; i < len(prevWindows); i = i + 1 {
			w1 := prevWindows[i]
			w2 := windows[i]

			if w1.Handle != w2.Handle {
				return true
			}
		}
		return false
	}
	for range ticker.C {
		windows := winapi.GetVisibleWindows()
		if anyChanged(windows) {
			nc.Publish(api.WindowsUpdated, utils.EncodeAny(windows))
		}
		prevWindows = windows
	}
}

func superFocusStealer(handle wintypes.HWND) wintypes.BOOL {
  // We should probably reset this...
	winapi.SystemParametersInfoA(wintypes.SPI_SETFOREGROUNDLOCKTIMEOUT, 0, 0, wintypes.SPIF_SENDCHANGE)
	success := winapi.SetForegroundWindow(handle)

	return success
}

func ListenIndefinitely() {
	nc, _ := nats.Connect(nats.DefaultURL)
	defer nc.Close()
	go poll(nc, time.Millisecond*300)
	nc.Subscribe(api.Windows, func(m *nats.Msg) {
		windows := winapi.GetVisibleWindows()
		m.Respond(utils.EncodeAny(windows))
	})
	nc.Subscribe(api.IsWindowFocused, func(m *nats.Msg) {
		window := utils.DecodeAny[wintypes.HWND](m.Data)
		current := winapi.GetForegroundWindow()
		focused := window == current
		response := utils.EncodeAny(focused)
		m.Respond(response)
	})
	nc.Subscribe(api.SetFocus, func(m *nats.Msg) {
		window := utils.DecodeAny[wintypes.HWND](m.Data)
		success := superFocusStealer(window)
		response := utils.EncodeAny(success)
		m.Respond(response)
	})
	// publish updates indefinitely
	select {}
}
