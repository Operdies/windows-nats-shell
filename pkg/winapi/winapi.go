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
	procGetMessage           = user32.MustFindProc("GetMessageW")

	setWinEventHook = user32.MustFindProc("SetWinEventHook")
	unhookWinEvent  = user32.MustFindProc("UnhookWinEvent")

	kernel             = syscall.MustLoadDLL("kernel32.dll")
	getCurrentThreadId = kernel.MustFindProc("GetCurrentThreadId")
	getModuleHandle    = kernel.MustFindProc("GetModuleHandleA")
	// loadLibrary        = kernel.MustFindProc("LoadLibrary")
	getProcAddress = kernel.MustFindProc("GetProcAddress")

	shell32       = syscall.MustLoadDLL("shell32.dll")
	shellExecuteW = shell32.MustFindProc("ShellExecuteA")
)

func ShellExecute(hwnd wintypes.HWND, lpOperation, lpFile, lpParameters, lpDirectory wintypes.LPCSTR, nShowCmd int) (wintypes.HINSTANCE, error) {
	r, _, err := shellExecuteW.Call(uintptr(hwnd), uintptr(lpOperation), uintptr(lpFile), uintptr(lpParameters), uintptr(lpDirectory), uintptr(nShowCmd))
  if r >= 32 {
    err = nil
  }
	return wintypes.HINSTANCE(r), err
}

func GetMessage(msg *wintypes.MSG, hwnd wintypes.HWND, msgFilterMin uint32, msgFilterMax uint32) int {
	ret, _, _ := procGetMessage.Call(
		uintptr(unsafe.Pointer(msg)),
		uintptr(hwnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax))
	return int(ret)
}
func GetProcAddress(hModule wintypes.HMODULE, lpProcName wintypes.LPCSTR) uintptr {
	r, _, _ := getProcAddress.Call(uintptr(hModule), uintptr(lpProcName))
	return r
}

// func LoadLibrary(lpLibFileName wintypes.LPCSTR) wintypes.HMODULE {
// 	r, _, _ := loadLibrary.Call(uintptr(lpLibFileName))
// 	return wintypes.HMODULE(r)
// }

func GetModuleHandleA(lpModuleName wintypes.LPCSTR) wintypes.HMODULE {
	res, _, _ := getModuleHandle.Call(uintptr(lpModuleName))
	return wintypes.HMODULE(res)
}

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
	// r0, _, e1 := syscall.SyscallN(getWindowTextW.Addr(), uintptr(hwnd), uintptr(unsafe.Pointer(str)), uintptr(maxCount))
	r0, _, e1 := getWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(str)), uintptr(maxCount))
	if r0 != 0 {
		err = e1
	}
	return
}

func IsWindowVisible(hwnd wintypes.HWND) bool {

	r0, _, _ := isWindowVisible.Call(uintptr(hwnd))
	return int32(r0) != 0
}

var allWindows struct {
	handles  []wintypes.HWND
	mut      sync.Mutex
	callback uintptr
}

func GetAllWindows() []wintypes.HWND {
	allWindows.mut.Lock()
	defer allWindows.mut.Unlock()
	allWindows.handles = make([]wintypes.HWND, 0)

	if allWindows.callback == 0 {
		allWindows.callback = syscall.NewCallback(
			func(h wintypes.HWND, p wintypes.LPARAM) wintypes.LRESULT {
				allWindows.handles = append(allWindows.handles, h)
				return 1
			})
	}
	EnumWindows(allWindows.callback, 0)
	return allWindows.handles
}

func GetVisibleWindows() []wintypes.Window {
	handles := GetAllWindows()
	result := make([]wintypes.Window, len(handles))
	b := make([]uint16, 200)
	k := 0
	focused := GetForegroundWindow()
	for _, h := range handles {
		if IsWindowVisible(h) {
			_, err := GetWindowText(h, &b[0], int32(len(b)))
			if err != nil {
				result[k] = wintypes.Window{Handle: h, Title: syscall.UTF16ToString(b), IsFocused: h == focused}
				k += 1
			}
		}
	}

	return result[:k]
}

func SetWindowsHookExW(idHook int, lpfn uintptr, hInstance wintypes.HINSTANCE, threadId wintypes.DWORD) wintypes.HHOOK {
	fmt.Printf("hInstance: %v\n", hInstance)
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

func UnhookWindowsHookEx(hhk wintypes.HHOOK) {
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
	defer UnhookWindowsHookEx(hook)

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
