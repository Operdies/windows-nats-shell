package wintypes

// SystemParametersInfo(SPI_SETFOREGROUNDLOCKTIMEOUT, 0, 0, SPIF_+UPDATEINIFILE)
const (
	WH_KEYBOARD_LL = 13
	TRUE           = 1
	FALSE          = 0

	SPI_SETFOREGROUNDLOCKTIMEOUT = 0x2001
	SPIF_UPDATEINIFILE           = 0x01
	SPIF_SENDCHANGE              = 0x02
	SPIF_SENDWININICHANGE        = 0x02
)

// https://docs.microsoft.com/en-us/windows/win32/winprog/windows-data-types
type (
	BOOL          int32
	BYTE          byte
	DWORD         uint32
	HANDLE        uintptr
	HHOOK         HANDLE
	HINSTANCE     HANDLE
	HMODULE       HANDLE
	HWND          HANDLE
	LONG          int32
	LONG_PTR      uintptr
	DWORD_PTR     uintptr
	LPARAM        LONG_PTR
	LRESULT       LONG_PTR
	PVOID         uintptr
	LPDWORD       DWORD_PTR
	WPARAM        uintptr
	HWINEVENTHOOK HANDLE
	PBYTE         []BYTE
	HOOKPROC      func(int, WPARAM, LPARAM) LRESULT
	WNDENUMPROC   func(HWND, LPARAM) LRESULT
	WINEVENTPROC  func(HWINEVENTHOOK, DWORD, HWND, LONG, LONG, DWORD, DWORD) uintptr
)

type MSG struct {
	Hwnd     HWND
	Message  uint32
	WParam   WPARAM
	LParam   LPARAM
	Time     DWORD
	Pt       POINT
	LPrivate DWORD
}

type POINT struct {
	X, Y LONG
}

type KBDLLHOOKSTRUCT struct {
	VkCode      DWORD
	ScanCode    DWORD
	Flags       DWORD
	Time        DWORD
	DwExtraInfo uintptr
}
