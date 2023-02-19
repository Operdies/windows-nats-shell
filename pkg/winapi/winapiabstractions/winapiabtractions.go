package winapiabstractions

import (
	"context"
	"fmt"
	"sort"
	"time"
	"unsafe"

	windows2 "github.com/operdies/windows-nats-shell/pkg/nats/api/windows"
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

func MoveWindow(hwnd wintypes.HWND, to windows2.Point) {
	winapi.SetWindowPos(hwnd, 0, int(to.X), int(to.Y), 0, 0, wintypes.SWP_NOACTIVATE|wintypes.SWP_NOOWNERZORDER|wintypes.SWP_NOSIZE)
}

func ResizeWindow(hwnd wintypes.HWND, cx, cy int) {
	winapi.SetWindowPos(hwnd, 0, 0, 0, cx, cy, wintypes.SWP_NOACTIVATE|wintypes.SWP_NOOWNERZORDER|wintypes.SWP_NOMOVE)
}

func SetWindowRect(hwnd wintypes.HWND, target windows2.Rect, resize bool) {
	styles := wintypes.SWP_NOACTIVATE
	styles |= wintypes.SWP_NOOWNERZORDER | wintypes.SWP_NOZORDER
	if !resize {
		styles |= wintypes.SWP_NOSIZE
	}
	if resize {
		// styles |= wintypes.SWP_ASYNCWINDOWPOS
	}
	winapi.SetWindowPos(hwnd, 0, int(target.Left), int(target.Top), int(target.Width()), int(target.Height()), uint(styles))
}

func SetZOrder(before, after wintypes.HWND) {
	styles := wintypes.SWP_NOACTIVATE | wintypes.SWP_NOSIZE | wintypes.SWP_NOMOVE
	winapi.SetWindowPos(after, before, 0, 0, 0, 0, styles)
}

func AnimateRectWithContext(hwnd wintypes.HWND, steps []windows2.Rect, ctx context.Context) {
	deadline, _ := ctx.Deadline()
	start := time.Now()
	timeLeft := deadline.Sub(start)
	each := timeLeft / time.Duration(len(steps))
	if each < 0 {
		each = time.Millisecond
	}
	ticker := time.NewTicker(each)

	timeDependentFrame := func() windows2.Rect {
		elapsed := time.Now().Sub(start)
		idx := elapsed / each
		if int(idx) >= len(steps) {
			return steps[len(steps)-1]
		}
		return steps[idx]
	}

	for {
		select {
		case <-ticker.C:
			SetWindowRect(hwnd, timeDependentFrame(), true)
		case <-ctx.Done():
			SetWindowRect(hwnd, steps[len(steps)-1], true)
			return
		}
	}
}

func MinimizeWindow(hwnd wintypes.HWND) {
	styles := uint64(winapi.GetWindowLong(hwnd, wintypes.GWL_STYLE))
	styles |= wintypes.WS_MINIMIZE
	winapi.SetWindowLongA(hwnd, wintypes.GWL_STYLE, wintypes.LONG(styles))
}

func RestoreWindow(hwnd wintypes.HWND) {
	styles := uint64(winapi.GetWindowLong(hwnd, wintypes.GWL_STYLE))
	styles &= ^wintypes.WS_MINIMIZE
	winapi.SetWindowLongA(hwnd, wintypes.GWL_STYLE, wintypes.LONG(styles))

}

func RestoreOrMinimize(h wintypes.HWND) {
	if WindowMinimized(h) == false {
		fmt.Printf("minimize: %v\n", h)
		MinimizeWindow(h)
	} else {
		fmt.Printf("restore: %v\n", h)
		RestoreWindow(h)
	}
	Redraw(h)

}

func WindowMinimized(hwnd wintypes.HWND) bool {
	styles := uint64(winapi.GetWindowLong(hwnd, wintypes.GWL_STYLE))
	return styles&wintypes.WS_MINIMIZE == wintypes.WS_MINIMIZE
}

// Save `value` in `m[`key`]` only if key does not exist
func maybeSave(m map[wintypes.HWND]uint64, value uint64, key wintypes.HWND) {
	if _, ok := m[key]; !ok {
		m[key] = value
	}
}

func HideBorder(hwnd wintypes.HWND) bool {
	styles := winapi.GetWindowLong(hwnd, wintypes.GWL_STYLE)
	exStyles := winapi.GetWindowLong(hwnd, wintypes.GWL_EXSTYLE)
	maybeSave(windowStyles, uint64(styles), hwnd)
	maybeSave(windowExStyles, uint64(exStyles), hwnd)
	styles &= ^int32(wintypes.WS_TILEDWINDOW)
	exStyles &= ^int32(wintypes.WS_EX_OVERLAPPEDWINDOW)
	winapi.SetWindowLongA(hwnd, wintypes.GWL_STYLE, wintypes.LONG(styles))
	winapi.SetWindowLongA(hwnd, wintypes.GWL_EXSTYLE, wintypes.LONG(exStyles))
	Redraw(hwnd)
	borderMap[hwnd] = false
	fmt.Printf("destroy: %v\n", hwnd)
	return true
}

func mkCallback() wintypes.WNDPROC {
	return func(h wintypes.HWND, u uint32, w wintypes.WPARAM, l wintypes.LPARAM) wintypes.LRESULT {
		if u == wintypes.WM_NCHITTEST {
			fmt.Printf("nowhere pog\n")
			return wintypes.LRESULT(wintypes.HTTRANSPARENT)
		}
		return winapi.DefWindowProcA(h, uint(u), w, l)
	}
}

func MakeClickThrough(hwnd wintypes.HWND) {
	exStyles := winapi.GetWindowLong(hwnd, wintypes.GWL_EXSTYLE)
	maybeSave(windowExStyles, uint64(exStyles), hwnd)
	exStyles |= int32(wintypes.WS_EX_LAYERED | wintypes.WS_EX_TRANSPARENT)
	winapi.SetWindowLongA(hwnd, wintypes.GWL_EXSTYLE, wintypes.LONG(exStyles))
	cb := windows.NewCallback(mkCallback())
	winapi.SetWindowLongPtrA(hwnd, int(wintypes.GWL_WNDPROC), cb)
}

func RestoreStyles(hwnd wintypes.HWND) bool {
	styles, ok := windowStyles[hwnd]
	if !ok {
		return false
	}
	exStyles, ok := windowExStyles[hwnd]
	if !ok {
		return false
	}
	winapi.SetWindowLongA(hwnd, wintypes.GWL_STYLE, wintypes.LONG(styles))
	winapi.SetWindowLongA(hwnd, wintypes.GWL_EXSTYLE, wintypes.LONG(exStyles))
	Redraw(hwnd)
	borderMap[hwnd] = true
	fmt.Printf("restore: %v\n", hwnd)
	return true
}

func ToggleBorder(hwnd wintypes.HWND) bool {
	if BordersEnabled(hwnd) {
		HideBorder(hwnd)
	} else {
		RestoreStyles(hwnd)
	}
	return true
}

var (
	borderMap      = map[wintypes.HWND]bool{}
	windowStyles   = map[wintypes.HWND]wintypes.WS_STYLES{}
	windowExStyles = map[wintypes.HWND]wintypes.WS_EX_STYLES{}
)

func BordersEnabled(h wintypes.HWND) bool {
	b, ok := borderMap[h]
	if ok {
		return b
	}

	return true
}

func Redraw(h wintypes.HWND) {
	winapi.SetWindowPos(h, 0, 0, 0, 0, 0, wintypes.SWP_FRAMECHANGED|wintypes.SWP_NOMOVE|wintypes.SWP_NOSIZE|wintypes.SWP_NOZORDER|wintypes.SWP_NOOWNERZORDER)
}

// var (
// 	BorderlessStyles   C.long = C.WS_CAPTION | C.WS_THICKFRAME | C.WS_MINIMIZEBOX | C.WS_MAXIMIZEBOX | C.WS_SYSMENU
// 	BorderlessExStyles C.long = 0 // C.WS_EX_DLGMODALFRAME | C.WS_EX_CLIENTEDGE | C.WS_EX_STATICEDGE
// 	RedrawFlags        C.uint = C.SWP_FRAMECHANGED | C.SWP_NOMOVE | C.SWP_NOSIZE | C.SWP_NOZORDER | C.SWP_NOOWNERZORDER
// )

// func EnableBorders(h wintypes.HWND) {
// 	hwnd := toCType(h)
// 	var lStyle C.long
// 	lStyle = C.GetWindowLong(hwnd, C.GWL_STYLE)
// 	lStyle |= BorderlessStyles
// 	C.SetWindowLong(hwnd, C.GWL_STYLE, lStyle)
//
// 	var eStyle C.long
// 	eStyle = C.GetWindowLong(hwnd, C.GWL_EXSTYLE)
// 	eStyle |= BorderlessExStyles
// 	C.SetWindowLong(hwnd, C.GWL_EXSTYLE, eStyle)
// 	redrawWindow(h)
// }
//
// func DisableBorders(h wintypes.HWND) {
// 	hwnd := toCType(h)
// 	lStyle := C.GetWindowLong(hwnd, C.GWL_STYLE)
// 	lStyle &= ^BorderlessStyles
// 	C.SetWindowLong(hwnd, C.GWL_STYLE, lStyle)
//
// 	eStyle := C.GetWindowLong(hwnd, C.GWL_EXSTYLE)
// 	eStyle &= ^BorderlessExStyles
// 	C.SetWindowLong(hwnd, C.GWL_EXSTYLE, eStyle)
// 	redrawWindow(h)
// }
//
// func redrawWindow(h wintypes.HWND) {
// 	hwnd := toCType(h)
// 	C.SetWindowPos(hwnd, nil, 0, 0, 0, 0, RedrawFlags)
// }
