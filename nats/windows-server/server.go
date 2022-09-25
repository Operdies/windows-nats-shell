// +build windows,amd64

package nats

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/operdies/minimalist-shell/nats/api"
	"github.com/operdies/minimalist-shell/winapi"
	"github.com/operdies/minimalist-shell/wintypes"
)

func mySelect[T1 any, T2 any](source []T1, selector func(T1) T2) []T2 {
	result := make([]T2, len(source))
	for i, item := range source {
		r := selector(item)
		result[i] = r
	}
	return result
}

func encodeAny[T any](value T) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(value)
	return buf.Bytes()
}

func decodeAny[T any](buffer []byte) T {
	reader := bytes.NewReader(buffer)
	dec := gob.NewDecoder(reader)
	var response T
	dec.Decode(&response)
	return response
}

func encodeWindows(windows []winapi.Window) []byte {
	bytes, _ := json.Marshal(windows)
	return bytes
}

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
			nc.Publish(api.WindowsUpdated, encodeWindows(windows))
		}
		prevWindows = windows
	}
}

/*
Windows set some countermeasures against focus stealing,
but apparently there are workarounds ?
*/
func superFocusStealer(handle wintypes.HWND) wintypes.BOOL {
	// Great, now all apps ever can steal focus whenever they want..
	winapi.SystemParametersInfoA(wintypes.SPI_SETFOREGROUNDLOCKTIMEOUT, 0, 0, wintypes.SPIF_SENDCHANGE)
	success := winapi.SetForegroundWindow(handle)
	// winapi.SystemParametersInfoA(wintypes.SPI_SETFOREGROUNDLOCKTIMEOUT, 0, 1000000, wintypes.SPIF_SENDCHANGE)

	return success
}

func ListenIndefinitely() {
	nc, _ := nats.Connect(nats.DefaultURL)
	defer nc.Close()
	go poll(nc, time.Millisecond*300)
	nc.Subscribe(api.Windows, func(m *nats.Msg) {
		windows := winapi.GetVisibleWindows()
		m.Respond(encodeWindows(windows))
	})
	nc.Subscribe(api.IsWindowFocused, func(m *nats.Msg) {
		window := decodeAny[wintypes.HWND](m.Data)
		current := winapi.GetForegroundWindow()
		focused := window == current
		response := encodeAny(focused)
		m.Respond(response)
	})
	nc.Subscribe(api.SetFocus, func(m *nats.Msg) {
		log.Println("Focus request!")
		window := decodeAny[wintypes.HWND](m.Data)
		success := superFocusStealer(window)
		fmt.Printf("success: %v\n", success)
		response := encodeAny(success)
		m.Respond(response)
	})
	// publish updates indefinitely
	select {}
}
