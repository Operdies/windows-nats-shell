//go:build windows && amd64
// +build windows,amd64

package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/operdies/windows-nats-shell/pkg/nats/client"
	"github.com/operdies/windows-nats-shell/pkg/winapi"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"
)

func poll(s client.Client, interval time.Duration) {
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
			s.PublishWindowsUpdated(windows)
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

func init() {
	go indexItems()
}

func getValues[T1 comparable, T2 any](source map[T1]T2) []T2 {
	result := make([]T2, len(source))
	k := 0
	for _, v := range source {
		result[k] = v
		k = k + 1
	}

	return result
}

func getKeys[T1 comparable, T2 any](source map[T1]T2) []T1 {
	result := make([]T1, len(source))
	k := 0
	for v := range source {
		result[k] = v
		k = k + 1
	}

	return result
}

func ListenIndefinitely() {
	client, _ := client.New(nats.DefaultURL)
	defer client.Close()
	go poll(client, time.Millisecond*1000)
	client.OnGetWindows(winapi.GetVisibleWindows)

	client.OnIsWindowFocused(func(h wintypes.HWND) bool {
		current := winapi.GetForegroundWindow()
		return current == h
	})

	client.OnSetFocus(func(h wintypes.HWND) bool {
		return superFocusStealer(h) == 1
	})

	client.OnGetPrograms(func() []string {
		// Ensure the files are properly indexed before proceeding
		indexItems()
		return getKeys(menuItems)
	})

	client.OnLaunchProgram(func(requested string) string {
		indexItems()
		if requested == "" {
			return "No program specified"
		}
		fmt.Println("Got command to start", requested)
		var err error
		val, ok := menuItems[requested]

		if ok {
			err = startDetachedProcess(val)
		} else {
      // As a fallback, just attempt to start it with 'cmd /c Start'
      // This makes sense for e.g. URLs or any other application which is able 
      // to handle some file extension or URI scheme
			err = startDetachedProcess(requested)
		}

		if err != nil {
			return err.Error()
		}
		return "Started " + requested
	})
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

func getPathItems() map[string]string {
	path, exists := os.LookupEnv("PATH")
	pathMap := map[string]string{}
	if exists {
		home, _ := os.LookupEnv("USERPROFILE")
		desktop := filepath.Join(home, "Desktop")
		for _, dir := range strings.Split(path+";"+desktop, ";") {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				// path/to/whatever does not exist
				continue
			}
			entries, _ := ioutil.ReadDir(dir)
			for _, path := range entries {
				nm := path.Name()
				ext := filepath.Ext(nm)
				if ext == ".exe" || ext == ".lnk" {
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
var indexMut sync.Mutex

func indexItems() {
	indexMut.Lock()
	defer indexMut.Unlock()
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

func main() {
	ListenIndefinitely()
}
