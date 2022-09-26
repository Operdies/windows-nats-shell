//go:build windows && amd64
// +build windows,amd64

package server

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/operdies/windows-nats-shell/pkg/nats/internal/api"
	"github.com/operdies/windows-nats-shell/pkg/nats/utils"
	"github.com/operdies/windows-nats-shell/pkg/winapi"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"
)

func poll(nc *nats.Conn, interval time.Duration) {
	ticker := time.NewTicker(interval)
	prevWindows := make([]wintypes.Window, 0)

	anyChanged := func(windows []wintypes.Window) bool {
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
  go IndexItems()
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
		log.Printf("Want to focus %v: %v\n", window, success)
		response := utils.EncodeAny(success)
		m.Respond(response)
	})
	nc.Subscribe(api.GetPrograms, func(m *nats.Msg) {
    IndexItems()
    response := utils.EncodeAny(menuItems)
		m.Respond(response)
	})
	// publish updates indefinitely
	select {}
}

type ProcStart struct {
	fullpath string
	args     []string
}

func getPathItems() []string {
	path, exists := os.LookupEnv("PATH")
	mymap := map[string]bool{}
	if exists {
		for _, dir := range strings.Split(path, ";") {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				// path/to/whatever does not exist
				continue
			}
			entries, _ := ioutil.ReadDir(dir)
			for _, path := range entries {
				if filepath.Ext(path.Name()) == ".exe" {
					mymap[path.Name()] = true
				}
			}
		}
	}
	executables := make([]string, 0)
	for k := range mymap {
		if len(k) > 0 {
			executables = append(executables, k)
		}
	}

	return executables
}

func getStartMenuItems() []string {
	// C:\Users\alexw\AppData\Roaming\Microsoft\Windows\Start Menu
	roaming := os.Getenv("APPDATA")
	startMenu := path.Join(roaming, "Microsoft", "Windows", "Start Menu")
	items := make([]string, 0)

	filepath.WalkDir(startMenu, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() == false && filepath.Ext(d.Name()) != ".ini" {
			items = append(items, path)
		}
		return nil
	})
	return items
}

var menuItems []string

func IndexItems() {
	if menuItems == nil {
		executables := getPathItems()
		executables = append(executables, getStartMenuItems()...)
    menuItems = executables
	}
}

func PublishPrograms() {
	nc, _ := nats.Connect(nats.DefaultURL)
	defer nc.Close()
  IndexItems()
	nc.Publish(api.GetPrograms, []byte(strings.Join(menuItems, "\n")))
}

func handler(hwinEventHook wintypes.HWINEVENTHOOK, event wintypes.DWORD, hwnd wintypes.HWND, idObject, idChild wintypes.LONG, idEventThread, dwmsEventTime wintypes.DWORD) uintptr {
	fmt.Printf("Got event!!!!!\n")
	return 0
}

func GetEvents() {
	fmt.Println("GetEvents")
	hook := winapi.Hooker(handler)
	fmt.Printf("hook: %v\n", hook)
	select {}
}
