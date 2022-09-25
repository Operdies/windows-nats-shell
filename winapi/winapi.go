// +build windows,amd64

package winapi

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"

	"github.com/operdies/minimalist-shell/wintypes"
)

var (
	user32 = syscall.MustLoadDLL("user32.dll")

	enumWindows              = user32.MustFindProc("EnumWindows")
	getWindowTextW           = user32.MustFindProc("GetWindowTextW")
	isWindowVisible          = user32.MustFindProc("IsWindowVisible")
	setWindowsHookExW        = user32.MustFindProc("SetWindowsHookExW")
	callNextHookEx           = user32.MustFindProc("CallNextHookEx")
	unhookWindowsHookEx      = user32.MustFindProc("UnhookWindowsHookEx")
	getForegroundWindow      = user32.MustFindProc("GetForegroundWindow")
	setForegroundWindow      = user32.MustFindProc("SetForegroundWindow")
	attachThreadInput        = user32.MustFindProc("AttachThreadInput")
	getWindowThreadProcessId = user32.MustFindProc("GetWindowThreadProcessId")
	systemParametersInfoA    = user32.MustFindProc("SystemParametersInfoA")

	kernel             = syscall.MustLoadDLL("kernel32.dll")
	getCurrentThreadId = kernel.MustFindProc("GetCurrentThreadId")
)

const (
	WH_CALLWNDPROC = 4
	WH_CBT         = 5
	WH_KEYBOARD    = 2
)

func EnumWindows(lpEnumFunc uintptr, lParam wintypes.LPARAM) (err error) {
	res, _, err := enumWindows.Call(
		lpEnumFunc,
		uintptr(lParam),
	)
	if res == 0 {
		err = nil
	}
	return
}

func GetForegroundWindow() wintypes.HWND {
	res, _, _ := getForegroundWindow.Call()
	return wintypes.HWND(res)
}

func SetForegroundWindow(hwnd wintypes.HWND) wintypes.BOOL {
	res, _, _ := setForegroundWindow.Call(uintptr(hwnd))
	return wintypes.BOOL(res)
}

func GetWindowText(hwnd wintypes.HWND, str *uint16, maxCount int32) (len int32, err error) {
	r0, _, e1 := syscall.SyscallN(getWindowTextW.Addr(), uintptr(hwnd), uintptr(unsafe.Pointer(str)), uintptr(maxCount))
	len = int32(r0)
	if len == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func IsWindowVisible(hwnd wintypes.HWND) bool {

	r0, _, _ := isWindowVisible.Call(uintptr(hwnd))
	return int32(r0) != 0
}

type Window struct {
	Title  string
	Handle wintypes.HWND
}

var vw struct {
	callback uintptr
	titles   []Window
	mut      sync.Mutex
}

func GetVisibleWindows() []Window {
	vw.mut.Lock()
	defer vw.mut.Unlock()
	vw.titles = make([]Window, 0)

	if vw.callback == 0 {
		cb := func(h wintypes.HWND, p wintypes.LPARAM) wintypes.LRESULT {
			b := make([]uint16, 200)
			_, err := GetWindowText(h, &b[0], int32(len(b)))
			if err != nil {
				// ignore and continue
				return 1
			}
			if IsWindowVisible(h) == false {
				return 1
			}
			title := syscall.UTF16ToString(b)

			vw.titles = append(vw.titles, Window{title, h})
			return 1
		}
		vw.callback = syscall.NewCallback(cb)
	}

	err := EnumWindows(vw.callback, 0)
	if err != nil {
		// fmt.Println(err.Error())
	}
	return vw.titles
}

func SetWindowsHookExW(idHook int, lpfn uintptr, hInstance uintptr, threadId int) uintptr {
	r0, _, err := setWindowsHookExW.Call(
		uintptr(idHook),
		lpfn,
		hInstance,
		uintptr(threadId),
	)
	if err != nil {
		fmt.Println(err.Error())
	}
	return r0
}

func CallNextHookEx(hhk wintypes.HHOOK, nCode int, wParam wintypes.WPARAM, lParam wintypes.LPARAM) wintypes.LRESULT {
	ret, _, _ := callNextHookEx.Call(
		uintptr(hhk),
		uintptr(nCode),
		uintptr(wParam),
		uintptr(lParam),
	)
	return wintypes.LRESULT(ret)
}

func AttachThreadInput(idAttach wintypes.DWORD, idAttachTo wintypes.DWORD, fAttach wintypes.BOOL) wintypes.BOOL {
	ret, _, _ := attachThreadInput.Call(uintptr(idAttach), uintptr(idAttachTo), uintptr(fAttach))
	return wintypes.BOOL(ret)
}

func GetWindowThreadProcessId(hwnd wintypes.HWND, lpdwProcessId wintypes.LPDWORD) wintypes.DWORD {
	ret, _, _ := getWindowThreadProcessId.Call(uintptr(hwnd), uintptr(lpdwProcessId))
	return wintypes.DWORD(ret)
}

func SystemParametersInfoA(uiAction uint, uiParam uint, pvParam wintypes.PVOID, fWinIni uint) wintypes.BOOL {
	r, _, _ := systemParametersInfoA.Call(uintptr(uiAction), uintptr(uiParam), uintptr(pvParam), uintptr(fWinIni))
	return wintypes.BOOL(r)
}

func UnhookWindowsHookEx(hhk int) {
	syscall.SyscallN(unhookWindowsHookEx.Addr(), uintptr(hhk))
}

func GetCurrentThreadId() wintypes.DWORD {
	r0, _, _ := syscall.SyscallN(getCurrentThreadId.Addr())
	return wintypes.DWORD(r0)
}

var cbt struct {
	callback    uintptr
	go_callback func(int)
	mut         sync.Mutex
}

func CBT(callback func(int)) {
	cbt.mut.Lock()
	defer cbt.mut.Unlock()

	cbt.go_callback = callback
	ch := make(chan bool)

	if cbt.callback == 0 {
		cb := func(ncode int, wparam wintypes.WPARAM, lparam wintypes.LPARAM) wintypes.LRESULT {
			if ncode >= 0 {
				ch <- true
				fmt.Println("New event!")
				cbt.go_callback(ncode)
				fmt.Printf("ncode: %v\n", ncode)
			}
			return CallNextHookEx(0, int(ncode), wparam, lparam)
		}
		cbt.callback = syscall.NewCallback(cb)
	}

	hook := SetWindowsHookExW(wintypes.WH_KEYBOARD_LL, cbt.callback, 0, 0)

	if hook == 0 {
		return
	}
	defer UnhookWindowsHookEx(int(hook))

	fmt.Println("Waiting for any event xD")
	<-ch
	fmt.Println("Got event. Exiting and unhooking.")
}
