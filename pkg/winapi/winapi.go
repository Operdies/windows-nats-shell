//go:build windows && amd64
// +build windows,amd64

package winapi

import (
	"log"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"

	windows2 "github.com/operdies/windows-nats-shell/pkg/nats/api/windows"
	"github.com/operdies/windows-nats-shell/pkg/winapi/internal/winapicgo"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"
)

var (
	user32 = windows.MustLoadDLL("user32.dll")

	setWindowPos             = user32.MustFindProc("SetWindowPos")
	enumWindows              = user32.MustFindProc("EnumWindows")
	getWindowRect            = user32.MustFindProc("GetWindowRect")
	getWindowTextW           = user32.MustFindProc("GetWindowTextW")
	isWindowVisible          = user32.MustFindProc("IsWindowVisible")
	setWindowsHookExA        = user32.MustFindProc("SetWindowsHookExA")
	callNextHookEx           = user32.MustFindProc("CallNextHookEx")
	unhookWindowsHookEx      = user32.MustFindProc("UnhookWindowsHookEx")
	getForegroundWindow      = user32.MustFindProc("GetForegroundWindow")
	setForegroundWindow      = user32.MustFindProc("SetForegroundWindow")
	showWindow               = user32.MustFindProc("ShowWindow")
	getAncestor              = user32.MustFindProc("GetAncestor")
	getWindow                = user32.MustFindProc("GetWindow")
	getParent                = user32.MustFindProc("GetParent")
	attachThreadInput        = user32.MustFindProc("AttachThreadInput")
	getWindowThreadProcessId = user32.MustFindProc("GetWindowThreadProcessId")
	systemParametersInfoA    = user32.MustFindProc("SystemParametersInfoA")
	getWindowLongA           = user32.MustFindProc("GetWindowLongA")
	setWindowLongA           = user32.MustFindProc("SetWindowLongA")
	setWindowLongPtrA        = user32.MustFindProc("SetWindowLongPtrA")

	defWindowProcA = user32.MustFindProc("DefWindowProcA")

	getMessageW      = user32.MustFindProc("GetMessageW")
	translateMessage = user32.MustFindProc("TranslateMessage")
	dispatchMessageW = user32.MustFindProc("DispatchMessageW")

	setWinEventHook = user32.MustFindProc("SetWinEventHook")
	unhookWinEvent  = user32.MustFindProc("UnhookWinEvent")

	kernel             = windows.MustLoadDLL("kernel32.dll")
	getCurrentThreadId = kernel.MustFindProc("GetCurrentThreadId")
	getModuleHandle    = kernel.MustFindProc("GetModuleHandleA")

	getProcAddress = kernel.MustFindProc("GetProcAddress")

	shell32       = windows.MustLoadDLL("shell32.dll")
	shellExecuteA = shell32.MustFindProc("ShellExecuteA")

	Shlwapi          = windows.MustLoadDLL("Shlwapi.dll")
	assocQueryString = Shlwapi.MustFindProc("AssocQueryStringA")
)

func SetWindowLongPtrA(hwnd wintypes.HWND, nIndex int, dwNewLong uintptr) uintptr {
	r, _, _ := setWindowLongPtrA.Call(uintptr(hwnd), uintptr(nIndex), dwNewLong)
	return r
}

func DefWindowProcA(hwnd wintypes.HWND, msg uint, wparam wintypes.WPARAM, lparam wintypes.LPARAM) wintypes.LRESULT {
	r, _, _ := defWindowProcA.Call(uintptr(hwnd), uintptr(msg), uintptr(wparam), uintptr(lparam))
	return wintypes.LRESULT(r)
}

func GetWindowLong(hwnd wintypes.HWND, index wintypes.GWL_INDEX) wintypes.GWL_INDEX {
	r, _, _ := getWindowLongA.Call(uintptr(hwnd), uintptr(index))
	return int32(r)
}

func SetWindowLongA(hwnd wintypes.HWND, index wintypes.GWL_INDEX, dwNewLong wintypes.LONG) (wintypes.LONG, error) {
	r, _, err := setWindowLongA.Call(uintptr(hwnd), uintptr(index), uintptr(dwNewLong))
	if r == 0 {
		return 0, err
	}
	return wintypes.LONG(r), nil
}

func SetWindowPos(hwnd wintypes.HWND, hwndInsertAfter wintypes.HWND, X, Y, cx, cy int, flags wintypes.SWP) error {
	r, _, err := setWindowPos.Call(uintptr(hwnd), uintptr(hwndInsertAfter), uintptr(X), uintptr(Y), uintptr(cx), uintptr(cy), uintptr(flags))
	if r != 0 {
		return err
	}
	return nil
}

func GetWindowRect(hwnd wintypes.HWND) (rect windows2.Rect) {
	getWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rect)))
	return rect
}

func GetAncestor(hwnd wintypes.HWND, gaFlags wintypes.GA_FLAGS) wintypes.HWND {
	parent, _, _ := getAncestor.Call(uintptr(hwnd), uintptr(gaFlags))
	return wintypes.HWND(parent)

}

func GetWindow(hwnd wintypes.HWND, uCmd wintypes.GW_CMD) wintypes.HWND {
	parent, _, _ := getWindow.Call(uintptr(hwnd), uintptr(uCmd))
	return wintypes.HWND(parent)
}

func GetParent(hwnd wintypes.HWND) wintypes.HWND {
	r, _, _ := getParent.Call(uintptr(hwnd))
	return wintypes.HWND(r)
}

func WindowFromPoint(point windows2.Point) wintypes.HWND {
	return winapicgo.WindowFromPoint(point)
}

func ShowWindow(hwnd wintypes.HWND, nCmdShow wintypes.N_CMD_SHOW) bool {
	r, _, _ := showWindow.Call(uintptr(hwnd), uintptr(nCmdShow))
	return r != 0
}
func AssocQueryString(flags wintypes.AssocF, str wintypes.AssocStr, pszAssoc, pszExtra wintypes.LPCSTR, pszOut wintypes.LPSTR, pcchOut uintptr) wintypes.HRESULT {
	r, _, _ := assocQueryString.Call(uintptr(flags), uintptr(str), uintptr(pszAssoc), uintptr(pszExtra), uintptr(pszOut), pcchOut)
	return wintypes.HRESULT(r)
}

func strPtr(s string) wintypes.LPCSTR {
	v := []byte(s)
	vv := unsafe.Pointer(&v[0])
	return wintypes.LPCSTR(vv)
}

func StartDetachedProcess(proc string, admin bool) error {
	const sw_shownormal = 1
	procPtr := strPtr(proc)
	var verbPtr wintypes.LPCSTR = 0
	verb := `open`
	if admin {
		verb = `runas`
		verbPtr = strPtr(verb)
	}
	_, err := ShellExecute(0, verbPtr, procPtr, 0, 0, sw_shownormal)
	return err
}

func ShellExecute(hwnd wintypes.HWND, lpOperation, lpFile, lpParameters, lpDirectory wintypes.LPCSTR, nShowCmd int) (wintypes.HINSTANCE, error) {
	r, _, err := shellExecuteA.Call(uintptr(hwnd), uintptr(lpOperation), uintptr(lpFile), uintptr(lpParameters), uintptr(lpDirectory), uintptr(nShowCmd))
	if r >= 32 {
		err = nil
	}
	return wintypes.HINSTANCE(r), err
}

func DispatchMessageW(msg **wintypes.MSG) wintypes.LRESULT {
	ret, _, _ := dispatchMessageW.Call(uintptr(unsafe.Pointer(msg)))
	return wintypes.LRESULT(ret)
}

func TranslateMessage(msg **wintypes.MSG) wintypes.BOOL {
	ret, _, _ := translateMessage.Call(uintptr(unsafe.Pointer(msg)))
	return wintypes.BOOL(ret)
}

func GetMessage(msg **wintypes.MSG, hwnd wintypes.HWND, msgFilterMin uint32, msgFilterMax uint32) int {
	ret, _, _ := getMessageW.Call(
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

func SetForegroundWindow(hwnd wintypes.HWND) (bool, error) {
	res, _, err := setForegroundWindow.Call(uintptr(hwnd))
	if res == 1 {
		return true, nil
	}
	return false, err
}

func GetWindowText(hwnd wintypes.HWND, str *uint16, maxCount int32) (len int32, err error) {
	// r0 is the number of copied characters
	r0, _, e1 := getWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(str)), uintptr(maxCount))
	len = int32(r0)
	if len > 0 {
		err = nil
	} else {
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
		allWindows.callback = windows.NewCallback(
			func(h wintypes.HWND, p wintypes.LPARAM) wintypes.LRESULT {
				allWindows.handles = append(allWindows.handles, h)
				return 1
			})
	}
	EnumWindows(allWindows.callback, 0)
	return allWindows.handles
}

func SetWindowsHookExW(idHook wintypes.WH_EVENTTYPE, lpfn uintptr, hInstance wintypes.HINSTANCE, threadId wintypes.DWORD) wintypes.HHOOK {
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

func UnhookWindowsHookEx(hhook wintypes.HHOOK) bool {
	r0, _, _ := unhookWindowsHookEx.Call(uintptr(hhook))
	success := r0 != 0
	return success
}
func UnhookWinEvent(hhook wintypes.HHOOK) bool {
	r0, _, _ := unhookWinEvent.Call(uintptr(hhook))
	success := r0 != 0
	return success
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

func GetCurrentThreadId() wintypes.DWORD {
	r0, _, _ := getCurrentThreadId.Call()
	return wintypes.DWORD(r0)
}

func SuperFocusStealer(handle wintypes.HWND) bool {
	SystemParametersInfoA(wintypes.SPI_SETFOREGROUNDLOCKTIMEOUT, 0, 0, wintypes.SPIF_SENDCHANGE)
	// it seems ShowWindow hides windows when explorer.exe isn't running, regardless of the argument verb
	success, _ := SetForegroundWindow(handle)

	return success
}
