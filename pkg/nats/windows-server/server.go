//go:build windows && amd64
// +build windows,amd64

package server

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
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
		data := make([]string, len(menuItems))
		i := 0
		for k := range menuItems {
			data[i] = k
			i = i + 1
		}
		response := utils.EncodeAny(data)
		m.Respond(response)
	})
	nc.Subscribe(api.LaunchProgram, func(m *nats.Msg) {
		IndexItems()
		requested := utils.DecodeAny[string](m.Data)
		fmt.Println("Got command to start", requested)
		var err error
		val, ok := menuItems[requested]
		// quote := "\""
		fmt.Println(requested, val, ok)
		if ok {
			err = startDetachedProcess(val)
		} else {
			err = startDetachedProcess(requested)
		}
		if err != nil {
			m.Respond([]byte(err.Error()))
		} else {
			m.Respond([]byte("Ok"))
		}
	})

	// publish updates indefinitely
	select {}
}

func startDetachedProcess(proc string) error {
	// We need to pass in some empty quotes so the start command can't misinterpret the first part of paths with spaces
	// as the window title
  cmd := exec.Command("cmd.exe")
  cmd.Args = nil
  // Forego any escaping because 'cmd /C start' is really particular
  cmd.SysProcAttr = &syscall.SysProcAttr{}
  cmd.SysProcAttr.CmdLine = `/C start "sos" "` + proc + `"`
	fmt.Println("Launching:", cmd.SysProcAttr.CmdLine)
  return cmd.Start()
}

type ProcStart struct {
	fullpath string
	args     []string
}

func getPathItems() map[string]string {
	path, exists := os.LookupEnv("PATH")
	pathMap := map[string]string{}
	if exists {
		for _, dir := range strings.Split(path, ";") {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				// path/to/whatever does not exist
				continue
			}
			entries, _ := ioutil.ReadDir(dir)
			for _, path := range entries {
				nm := path.Name()
				if filepath.Ext(nm) == ".exe" {
					pathMap[baseNameNoExt(nm)] = filepath.Join(dir, nm)
				}
			}
		}
	}

	return pathMap
}

func baseNameNoExt(fullname string) string {
	bn := filepath.Base(fullname)
	idx := strings.LastIndex(bn, ".")
	if idx > 0 {
		return bn[:idx]
	}
	return bn
}

func getStartMenuItems() map[string]string {
	// C:\Users\alexw\AppData\Roaming\Microsoft\Windows\Start Menu
	roaming := os.Getenv("APPDATA")
	startMenu := path.Join(roaming, "Microsoft", "Windows", "Start Menu")
	items := map[string]string{}

	filepath.WalkDir(startMenu, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() == false && filepath.Ext(d.Name()) != ".ini" {
			nm := d.Name()
			items[baseNameNoExt(nm)] = path
		}
		return nil
	})
	return items
}

var menuItems map[string]string

func IndexItems() {
	if menuItems == nil {
		executables := getPathItems()
		for k, v := range getStartMenuItems() {
			executables[k] = v
		}
		menuItems = executables
	}
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
