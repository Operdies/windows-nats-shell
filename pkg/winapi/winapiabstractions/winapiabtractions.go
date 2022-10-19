package winapiabstractions

import (
	"sort"
	"unsafe"

	// natswindows "github.com/operdies/windows-nats-shell/pkg/nats/api/windows"
	"github.com/operdies/windows-nats-shell/pkg/winapi"
	"github.com/operdies/windows-nats-shell/pkg/winapi/internal/winapicgo"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"
	"golang.org/x/sys/windows"
)

func GetSiblings(hwnd wintypes.HWND) (prev, next wintypes.HWND) {
	windows := GetVisibleWindows()
	for i, w := range windows {
		if w.Title == "Background" {
			windows = append(windows[:i], windows[i+1:]...)
			break
		}
	}
	if len(windows) == 0 {
		return
	}
	sort.Slice(windows, func(i, j int) bool {
		return int(windows[i].Handle) < int(windows[j].Handle)
	})

	for i, h := range windows {
		if h.Handle == hwnd {
			if i == 0 {
				prev = windows[len(windows)-1].Handle
			} else {
				prev = windows[i-1].Handle
			}
			if i == len(windows)-1 {
				next = windows[0].Handle
			} else {
				next = windows[i+1].Handle
			}
			break
		}
	}
	if prev == 0 {
		prev = windows[0].Handle
	}
	if next == 0 {
		next = windows[len(windows)-1].Handle
	}
	return
}

func GetWindowTextEasy(h wintypes.HWND) (str string, err error) {
	b := make([]uint16, 200)
	_, err = winapi.GetWindowText(h, &b[0], int32(len(b)))
	if err != nil {
		return "", err
	}
	str = windows.UTF16ToString(b)
	return str, nil
}

func GetVisibleWindows() []wintypes.Window {
	handles := winapi.GetAllWindows()
	result := make([]wintypes.Window, len(handles))
	k := 0
	focused := winapi.GetForegroundWindow()
	for i, h := range handles {
		if winapi.IsWindowVisible(h) {
			title, err := GetWindowTextEasy(h)
			if err == nil {
				result[k] = wintypes.Window{Handle: h, Title: title, IsFocused: h == focused, ZOrder: i}
				k += 1
			}
		}
	}

	return result[:k]
}

// Tool windows don't appear in the app switcher or the task bar
func MakeToolWindow(hwnd unsafe.Pointer) {
	winapicgo.MakeToolWindow(hwnd)
}

func SetBottomMost(hwnd unsafe.Pointer) {
	winapicgo.SetBottomMost(hwnd)
}

func MoveWindow(hwnd wintypes.HWND, to wintypes.POINT) {
	winapi.SetWindowPos(hwnd, 0, int(to.X), int(to.Y), 0, 0, wintypes.SWP_NOACTIVATE|wintypes.SWP_NOOWNERZORDER|wintypes.SWP_NOSIZE)
}

func ResizeWindow(hwnd wintypes.HWND, cx, cy int) {
	winapi.SetWindowPos(hwnd, 0, 0, 0, cx, cy, wintypes.SWP_NOACTIVATE|wintypes.SWP_NOOWNERZORDER|wintypes.SWP_NOMOVE)
}

func SetWindowRect(hwnd wintypes.HWND, target wintypes.RECT) {
	l := int(target.Left)
	t := int(target.Top)
	r := int(target.Right)
	b := int(target.Bottom)
	width := r - l
	height := b - t
	winapi.SetWindowPos(hwnd, 0, l, t, width, height, wintypes.SWP_NOACTIVATE|wintypes.SWP_NOOWNERZORDER)
}

func HideBorder(hwnd wintypes.HWND) bool {
	winapicgo.DisableBorders(hwnd)
	return true
}

func ShowBorder(hwnd wintypes.HWND) bool {
	winapicgo.EnableBorders(hwnd)
	return true
}

func ToggleBorder(hwnd wintypes.HWND) bool {
	winapicgo.ToggleBorders(hwnd)
	return true
}
