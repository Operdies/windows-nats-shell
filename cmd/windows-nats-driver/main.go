//go:build windows && amd64

package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unsafe"

	"github.com/nats-io/nats.go"

	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
	"github.com/operdies/windows-nats-shell/pkg/utils/files"
	"github.com/operdies/windows-nats-shell/pkg/winapi"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"
)

type executableSource struct {
	Path      string
	Recursive bool
	Watch     bool
}

type launcherOptions struct {
	IncludeSystemPath bool
	WatchSystemPath   bool
	Sources           []executableSource
}

type customOptions struct {
	Launcher launcherOptions
}

func superFocusStealer(handle wintypes.HWND) wintypes.BOOL {
	// We should probably reset this...
	winapi.SystemParametersInfoA(wintypes.SPI_SETFOREGROUNDLOCKTIMEOUT, 0, 0, wintypes.SPIF_SENDCHANGE)
	success := winapi.SetForegroundWindow(handle)

	return success
}

func ListenIndefinitely() {
	client, _ := client.New(nats.DefaultURL, time.Second)
	defer client.Close()

	cfg, err := client.Request.Config("")
	if err != nil {
		panic(err)
	}
	custom, _ := shell.GetCustom[customOptions](cfg)
	indexItems(custom)

	client.Subscribe.ShellEvent(func(e shell.Event) {
		if e.NCode == shell.HSHELL_ACTIVATESHELLWINDOW ||
			e.NCode == shell.HSHELL_WINDOWDESTROYED ||
			e.NCode == shell.HSHELL_WINDOWREPLACED ||
			e.NCode == shell.HSHELL_WINDOWACTIVATED ||
			e.NCode == shell.HSHELL_WINDOWCREATED {
			client.Publish.WindowsUpdated(winapi.GetVisibleWindows())
		}
	})
	client.Subscribe.GetWindows(winapi.GetVisibleWindows)

	client.Subscribe.IsWindowFocused(func(h wintypes.HWND) bool {
		current := winapi.GetForegroundWindow()
		return current == h
	})

	client.Subscribe.SetFocus(func(h wintypes.HWND) bool {
		return superFocusStealer(h) == 1
	})

	client.Subscribe.GetPrograms(func() []string {
		// Ensure the files are properly indexed before proceeding
		return getFriendlyNames()
	})

	client.Subscribe.LaunchProgram(func(requested string) string {
		if requested == "" {
			return "No program specified"
		}
		fmt.Println("Got command to start", requested)
		val, err := getPathExecutable(requested)

		if err != nil {
			return err.Error()
		}

		err = startDetachedProcess(val)

		if err != nil {
			return err.Error()
		}
		return "Started " + requested
	})
	select {}
}

func startDetachedProcess(proc string) error {
	const sw_shownormal = 1

	// b := make([]byte, 0)
	empty := 0
	procb := []byte(proc)
	procp := unsafe.Pointer(&procb[0])
	_, err := winapi.ShellExecute(0, wintypes.LPCSTR(empty), wintypes.LPCSTR(procp), wintypes.LPCSTR(empty), wintypes.LPCSTR(empty), sw_shownormal)
	// shellExecute(0, "", proc, "", "", sw_shownormal)
	return err
}

func mergeMaps(maps ...map[string]string) map[string]string {
	result := map[string]string{}
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
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
func getFriendlyNames() []string {
	maps := make([]map[string]string, 0, len(watchers))
	for _, w := range watchers {
		maps = append(maps, w.Files())
	}

	// Reverse the maps array. mergeMaps overwrites
	// the keys whenever a new key is found, but
	// launchProgram uses the first match it finds.
	// This is to ensure that the selected program
	// actually matches the program which is executed.
	// This probably doesn't matter since only
	// the friendly name is presented anyway (rigth now)
	for i, j := 0, len(maps)-1; i < j; i, j = i+1, j-1 {
		maps[i], maps[j] = maps[j], maps[i]
	}

	merged := mergeMaps(maps...)
	return getKeys(merged)
}

func getPathExecutable(s string) (prog string, err error) {
	for _, w := range watchers {
		if p, ok := w.Files()[s]; ok {
			prog = p
			return
		}
	}
	err = fmt.Errorf("File %s not found.", s)
	return
}

var (
	// menuItems map[string]string
	// indexMut  sync.Mutex
	watchers = make([]*files.WatchedDir, 0, 20)
)

func indexItems(custom customOptions) {
	for _, source := range custom.Launcher.Sources {
		watchers = append(watchers, files.Create(source.Path, source.Recursive, source.Watch))
	}

	if custom.Launcher.IncludeSystemPath {
		path, exists := os.LookupEnv("PATH")
		if exists {
			for _, dir := range strings.Split(path, ";") {
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					// path/to/whatever does not exist
					continue
				}
				watchers = append(watchers, files.Create(dir, false, custom.Launcher.WatchSystemPath))
			}
		}
	}
}

func main() {
	ListenIndefinitely()
}
