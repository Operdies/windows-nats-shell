//go:build windows && amd64
// +build windows,amd64

package winapi

import (
	"fmt"
	"log"
	"sync"
	"syscall"
	"unsafe"

	"github.com/operdies/windows-nats-shell/pkg/wintypes"
)

var (
	user32 = syscall.MustLoadDLL("user32.dll")

	enumWindows              = user32.MustFindProc("EnumWindows")
	getWindowTextW           = user32.MustFindProc("GetWindowTextW")
	isWindowVisible          = user32.MustFindProc("IsWindowVisible")
	setWindowsHookExA        = user32.MustFindProc("SetWindowsHookExA")
	callNextHookEx           = user32.MustFindProc("CallNextHookEx")
	unhookWindowsHookEx      = user32.MustFindProc("UnhookWindowsHookEx")
	getForegroundWindow      = user32.MustFindProc("GetForegroundWindow")
	setForegroundWindow      = user32.MustFindProc("SetForegroundWindow")
	attachThreadInput        = user32.MustFindProc("AttachThreadInput")
	getWindowThreadProcessId = user32.MustFindProc("GetWindowThreadProcessId")
	systemParametersInfoA    = user32.MustFindProc("SystemParametersInfoA")

	setWinEventHook = user32.MustFindProc("SetWinEventHook")
	unhookWinEvent  = user32.MustFindProc("UnhookWinEvent")

	kernel             = syscall.MustLoadDLL("kernel32.dll")
	getCurrentThreadId = kernel.MustFindProc("GetCurrentThreadId")
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

var vw struct {
	callback      uintptr
	titles        []wintypes.Window
	mut           sync.Mutex
	focusedWindow wintypes.HWND
}

func GetVisibleWindows() []wintypes.Window {
	vw.mut.Lock()
	defer vw.mut.Unlock()
	vw.titles = make([]wintypes.Window, 0)
	vw.focusedWindow = GetForegroundWindow()

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

			vw.titles = append(vw.titles, wintypes.Window{Title: title, Handle: h, IsFocused: h == vw.focusedWindow})
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

func SetWindowsHookExW(idHook int, lpfn uintptr, hInstance wintypes.HINSTANCE, threadId wintypes.DWORD) wintypes.HHOOK {
	r0, _, err := setWindowsHookExA.Call(
		uintptr(idHook),
		lpfn,
		uintptr(hInstance),
		uintptr(threadId),
	)

	if r0 == 0 {
		log.Fatal(err.Error())
	}
	return wintypes.HHOOK(r0)
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

func CBT(callback func(int)) {
	var cbt struct {
		callback    uintptr
		go_callback func(int)
	}

	cbt.go_callback = callback
	var wg sync.WaitGroup
	wg.Add(1)

	if cbt.callback == 0 {
		var cb wintypes.HOOKPROC
		cb = func(ncode int, wparam wintypes.WPARAM, lparam wintypes.LPARAM) wintypes.LRESULT {
			fmt.Println("New event!")
			cbt.go_callback(ncode)
			fmt.Printf("ncode: %v\n", ncode)
			wg.Done()
			return CallNextHookEx(0, int(ncode), wparam, lparam)
		}
		cbt.callback = syscall.NewCallback(cb)
	}

	hook := SetWindowsHookExW(wintypes.WH_SHELL, cbt.callback, 0, GetCurrentThreadId())

	if hook == 0 {
		log.Fatal("Hook = 0")
		return
	}
	defer UnhookWindowsHookEx(int(hook))

	fmt.Println("Waiting for any event xD")
	wg.Wait()
	fmt.Println("Got event. Exiting and unhooking.")
}

func SetWinEventHook(eventMin, eventMax wintypes.DWORD,
	hmodWinEventProc wintypes.HMODULE,
	pfnWinEventProc wintypes.WINEVENTPROC,
	idProcess, idThread, dwFlags wintypes.DWORD) wintypes.HWINEVENTHOOK {
	hhook, _, err := setWinEventHook.Call(uintptr(eventMin),
		uintptr(eventMax),
		uintptr(hmodWinEventProc),
		syscall.NewCallback(pfnWinEventProc),
		uintptr(idProcess),
		uintptr(idThread),
		uintptr(dwFlags))

	if hhook == 0 {
		log.Fatal(err)
	}
	return wintypes.HWINEVENTHOOK(hhook)
}

func Hooker(fn wintypes.WINEVENTPROC) wintypes.HWINEVENTHOOK {
	 min, max := 0x00000001, 0x7FFFFFFF
	return SetWinEventHook(wintypes.DWORD(min), wintypes.DWORD(max), 0, fn, 0, 0, wintypes.WINEVENT_OUTOFCONTEXT)
}
