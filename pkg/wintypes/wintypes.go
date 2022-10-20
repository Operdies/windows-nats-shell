package wintypes

import (
	"math"
	"unsafe"
)

// SystemParametersInfo(SPI_SETFOREGROUNDLOCKTIMEOUT, 0, 0, SPIF_+UPDATEINIFILE)

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
	LPSTR         uintptr
	LPCSTR        uintptr
	PVOID         uintptr
	LPDWORD       DWORD_PTR
	WPARAM        uintptr
	HWINEVENTHOOK HANDLE
	PBYTE         []BYTE
	WNDPROC       func(HWND, uint32, WPARAM, LPARAM) LRESULT
	HOOKLLPROC    func(int, WPARAM, unsafe.Pointer) LRESULT
	HOOKPROC      func(int32, WPARAM, LPARAM) LRESULT
	WNDENUMPROC   func(HWND, LPARAM) LRESULT
	WINEVENTPROC  func(HWINEVENTHOOK, DWORD, HWND, LONG, LONG, DWORD, DWORD) uintptr

	AssocF   int32
	AssocStr int32
	HRESULT  LONG
	LWSTDAPI HRESULT

	WH_EVENTTYPE int
	N_CMD_SHOW   int
)

type GWL_INDEX = int32

const (

	// Sets a new extended window style.
	GWL_EXSTYLE GWL_INDEX = -20
	// Sets a new application instance handle.
	GWL_HINSTANCE = -6
	// Sets a new identifier of the child window. The window cannot be a top-level window.
	GWL_ID = -12
	// Sets a new window style.
	GWL_STYLE = -16
	// Sets the user data associated with the window. This data is intended for use by the application that created the window. Its value is initially zero.
	GWL_USERDATA = -21
	// Sets a new address for the window procedure.
	// You cannot change this attribute if the window does not belong to the same process as the calling thread.
	GWL_WNDPROC = -4
)

const (
	SW_HIDE            N_CMD_SHOW = 0
	SW_SHOWNORMAL                 = 1
	SW_NORMAL                     = 1
	SW_SHOWMINIMIZED              = 2
	SW_SHOWMAZIMIZED              = 3
	SW_MAZIMIZED                  = 3
	SW_SHOWNOACTIVATE             = 4
	SW_SHOW                       = 5
	SW_MINIMIZE                   = 6
	SW_SHOWMINNOACTIVE            = 7
	SW_SHOWNA                     = 8
	SW_RESTORE                    = 9
	SW_SHOWDEFAULT                = 10
	SW_FORCEMINIMIZE              = 11
)

const (
	WH_KEYBOARD    WH_EVENTTYPE = 2
	WH_KEYBOARD_LL              = 13
	WH_MOUSE_LL                 = 14
	WH_CALLWNDPROC              = 4
	WH_SHELL                    = 10
)

const (
	SPI_SETFOREGROUNDLOCKTIMEOUT = 0x2001
	SPIF_UPDATEINIFILE           = 0x01
	SPIF_SENDCHANGE              = 0x02
	SPIF_SENDWININICHANGE        = 0x02
)

const (
	WINEVENT_OUTOFCONTEXT = 0
	WINEVENT_INCONTEXT    = 4
	TRUE                  = 1
	FALSE                 = 0
)

type SWP = uint

const (
	//	If the calling thread and the thread that owns the window are attached to different input queues, the system posts the request to the thread that owns the window. This prevents the calling thread from blocking its execution while other threads process the request.
	SWP_ASYNCWINDOWPOS SWP = 0x4000

	//	Prevents generation of the WM_SYNCPAINT message.
	SWP_DEFERERASE = 0x2000

	//	Draws a frame (defined in the window's class description) around the window.
	SWP_DRAWFRAME = 0x0020

	//	Applies new frame styles set using the SetWindowLong function. Sends a WM_NCCALCSIZE message to the window, even if the window's size is not being changed. If this flag is not specified, WM_NCCALCSIZE is sent only when the window's size is being changed.
	SWP_FRAMECHANGED = 0x0020

	//	Hides the window.
	SWP_HIDEWINDOW = 0x0080

	//	Does not activate the window. If this flag is not set, the window is activated and moved to the top of either the topmost or non-topmost group (depending on the setting of the hWndInsertAfter parameter).
	SWP_NOACTIVATE = 0x0010

	//	Discards the entire contents of the client area. If this flag is not specified, the valid contents of the client area are saved and copied back into the client area after the window is sized or repositioned.
	SWP_NOCOPYBITS = 0x0100

	//	Retains the current position (ignores X and Y parameters).
	SWP_NOMOVE = 0x0002

	//	Does not change the owner window's position in the Z order.
	SWP_NOOWNERZORDER = 0x0200

	//	Does not redraw changes. If this flag is set, no repainting of any kind occurs. This applies to the client area, the nonclient area (including the title bar and scroll bars), and any part of the parent window uncovered as a result of the window being moved. When this flag is set, the application must explicitly invalidate or redraw any parts of the window and parent window that need redrawing.
	SWP_NOREDRAW = 0x0008

	//	Same as the SWP_NOOWNERZORDER flag.
	SWP_NOREPOSITION = 0x0200

	//	Prevents the window from receiving the WM_WINDOWPOSCHANGING message.
	SWP_NOSENDCHANGING = 0x0400

	//	Retains the current size (ignores the cx and cy parameters).
	SWP_NOSIZE = 0x0001

	//	Retains the current Z order (ignores the hWndInsertAfter parameter).
	SWP_NOZORDER = 0x0004

	//	Displays the window.
	SWP_SHOWWINDOW = 0x0040
)

func SUCCEEDED(code HRESULT) bool {
	// The highest bit of an HRESULT indicates success
	// If it is set to 0 it indicates success. Otherwise the object contains a failure code
	thirtyFirst := (code >> 30) == 0
	return thirtyFirst
}

const (
	ASSOCF_NONE                 AssocF = 0x00000000
	ASSOCF_INIT_NOREMAPCLSID           = 0x00000001
	ASSOCF_INIT_BYEXENAME              = 0x00000002
	ASSOCF_OPEN_BYEXENAME              = 0x00000002
	ASSOCF_INIT_DEFAULTTOSTAR          = 0x00000004
	ASSOCF_INIT_DEFAULTTOFOLDER        = 0x00000008
	ASSOCF_NOUSERSETTINGS              = 0x00000010
	ASSOCF_NOTRUNCATE                  = 0x00000020
	ASSOCF_VERIFY                      = 0x00000040
	ASSOCF_REMAPRUNDLL                 = 0x00000080
	ASSOCF_NOFIXUPS                    = 0x00000100
	ASSOCF_IGNOREBASECLASS             = 0x00000200
	ASSOCF_INIT_IGNOREUNKNOWN          = 0x00000400
	ASSOCF_INIT_FIXED_PROGID           = 0x00000800
	ASSOCF_IS_PROTOCOL                 = 0x00001000
)

const (
	ASSOCSTR_COMMAND AssocStr = 1
	ASSOCSTR_EXECUTABLE
	ASSOCSTR_FRIENDLYDOCNAME
	ASSOCSTR_FRIENDLYAPPNAME
	ASSOCSTR_NOOPEN
	ASSOCSTR_SHELLNEWVALUE
	ASSOCSTR_DDECOMMAND
	ASSOCSTR_DDEIFEXEC
	ASSOCSTR_DDEAPPLICATION
	ASSOCSTR_DDETOPIC
	ASSOCSTR_INFOTIP
	ASSOCSTR_QUICKTIP
	ASSOCSTR_TILEINFO
	ASSOCSTR_CONTENTTYPE
	ASSOCSTR_DEFAULTICON
	ASSOCSTR_SHELLEXTENSION
	ASSOCSTR_DROPTARGET
	ASSOCSTR_DELEGATEEXECUTE
	ASSOCSTR_SUPPORTED_URI_PROTOCOLS
	ASSOCSTR_PROGID
	ASSOCSTR_APPID
	ASSOCSTR_APPPUBLISHER
	ASSOCSTR_APPICONREFERENCE
	ASSOCSTR_MAX
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

type Window struct {
	Title     string
	Handle    HWND
	IsFocused bool
	ZOrder    int
}

type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

func (r RECT) Transform(x, y int32) RECT {
	return RECT{
		Left:   r.Left + x,
		Right:  r.Right + x,
		Top:    r.Top + y,
		Bottom: r.Bottom + y,
	}
}

func (r RECT) Width() int32 {
	return r.Right - r.Left
}

func (r RECT) Height() int32 {
	return r.Bottom - r.Top
}

func (r RECT) CenterAround(other POINT) RECT {
	rect := RECT{
		Left:   int32(other.X),
		Right:  int32(other.X),
		Top:    int32(other.Y),
		Bottom: int32(other.Y),
	}
	return r.CenterIn(rect)
}

func animate(start, end float64, steps int) []float64 {
	ret := make([]float64, 0, steps)
	step := (end - start) / float64(steps)
	for i := 0; i < steps; i++ {
		ret = append(ret, start+(step*float64(i+1)))
	}
	return ret
}

// Animate rect `r` towards rect `to` with `frames` frames
func (r RECT) Animate(final RECT, frames int, animateSize bool) []RECT {
	to := final
	if !animateSize {
		to = r.CenterIn(to)
	}
	lefts := animate(float64(r.Left), float64(to.Left), frames)
	rights := animate(float64(r.Right), float64(to.Right), frames)
	tops := animate(float64(r.Top), float64(to.Top), frames)
	bottoms := animate(float64(r.Bottom), float64(to.Bottom), frames)
	result := make([]RECT, 0, frames)
	for i := 0; i < frames; i++ {
		result = append(result, RECT{
			Left:   int32(lefts[i]),
			Right:  int32(rights[i]),
			Top:    int32(tops[i]),
			Bottom: int32(bottoms[i]),
		})
	}
	result[len(result)-1] = final
	return result
}

// Given a number between 0 and 1, get a corresponding point
// on the rectangle perimeter such that 0 and 1 is the center of the top line,
// 0 and 1 is the center of the top line
// 0.125 is the top right corner
// 0.375 is the bottom right corner
// 0.625 is the bottom left corner
// 0.875 is the top left corner
func (r RECT) GetPointOnPerimeter(point float64) POINT {
	// put point in the range 0 <= point < 1
	point = point - math.Floor(point)
	rad := point * math.Pi * 2
	sin := math.Sin(rad)
	cos := math.Cos(rad)

	// Extrapolate the point on the circle onto the closest side of the surrounding square
	factor := 1.0 / math.Max(math.Abs(cos), math.Abs(sin))
	cos *= factor
	sin *= factor

	return POINT{
		X: LONG(r.Left) + (LONG(r.Width()/2) + LONG(math.Round(sin*float64(r.Width()/2)))),
		Y: LONG(r.Top) + (LONG(r.Height()/2) - LONG(math.Round(cos*float64(r.Height()/2)))),
	}
}

func (r RECT) CenterIn(other RECT) RECT {
	otherHalfWidth := other.Width() / 2
	rHalfWidth := r.Width() / 2
	otherHalfHeight := other.Height() / 2
	rHalfHeight := r.Height() / 2
	return RECT{
		Left:   other.Left + (otherHalfWidth - rHalfWidth),
		Right:  other.Left + (otherHalfWidth + rHalfWidth),
		Top:    other.Top + (otherHalfHeight - rHalfHeight),
		Bottom: other.Top + (otherHalfHeight + rHalfHeight),
	}
}

func (r RECT) Scale(factor float64) RECT {
	centered := r.CenterIn(RECT{0, 0, 0, 0})
	centered.Left = int32(factor * float64(centered.Left))
	centered.Right = int32(factor * float64(centered.Right))
	centered.Top = int32(factor * float64(centered.Top))
	centered.Bottom = int32(factor * float64(centered.Bottom))
	return centered.CenterIn(r)
}

type POINT struct {
	X, Y LONG
}

func (p1 POINT) DistanceTo(p2 POINT) float64 {
	x1 := float64(p1.X)
	x2 := float64(p2.X)
	y1 := float64(p1.Y)
	y2 := float64(p2.Y)
	return math.Sqrt(math.Pow(x1-x2, 2) + math.Pow(y1-y2, 2))
}
func (p POINT) Add(p2 POINT) POINT {
	return POINT{X: p.X + p2.X, Y: p.Y + p2.Y}
}
func (p POINT) Sub(p2 POINT) POINT {
	return POINT{X: p.X - p2.X, Y: p.Y - p2.Y}
}

type GW_CMD = uint

const (
	GW_HWNDFIRST    GW_CMD = 0
	GW_HWNDLAST            = 1
	GW_HWNDNEXT            = 2
	GW_HWNDPREV            = 3
	GW_OWNER               = 4
	GW_CHILD               = 5
	GW_ENABLEDPOPUP        = 6
)

type GA_FLAGS = uint

const (
	GA_PARENT    GA_FLAGS = 1
	GA_ROOT               = 2
	GA_ROOTOWNER          = 3
)

type WS_STYLES = uint64

const (
	// The window has a thin-line border
	WS_BORDER = 0x00800000
	// The window has a title bar (includes the WS_BORDER style).
	WS_CAPTION = 0x00C00000
	// The window is a child window. A window with this style cannot have a menu bar. This style cannot be used with the WS_POPUP style.
	WS_CHILD = 0x40000000
	// Same as the WS_CHILD style.
	WS_CHILDWINDOW = 0x40000000
	// Excludes the area occupied by child windows when drawing occurs within the parent window. This style is used when creating the parent window.
	WS_CLIPCHILDREN = 0x02000000
	// Clips child windows relative to each other; that is, when a particular child window receives a WM_PAINT message, the WS_CLIPSIBLINGS style clips all other overlapping child windows out of the region of the child window to be updated. If WS_CLIPSIBLINGS is not specified and child windows overlap, it is possible, when drawing within the client area of a child window, to draw within the client area of a neighboring child window.
	WS_CLIPSIBLINGS = 0x04000000
	// The window is initially disabled. A disabled window cannot receive input from the user. To change this after a window has been created, use the EnableWindow function.
	WS_DISABLED = 0x08000000
	// The window has a border of a style typically used with dialog boxes. A window with this style cannot have a title bar.
	WS_DLGFRAME = 0x00400000
	// The window is the first control of a group of controls. The group consists of this first control and all controls defined after it, up to the next control with the WS_GROUP style. The first control in each group usually has the WS_TABSTOP style so that the user can move from group to group. The user can subsequently change the keyboard focus from one control in the group to the next control in the group by using the direction keys. You can turn this style on and off to change dialog box navigation. To change this style after a window has been created, use the SetWindowLong function.
	WS_GROUP = 0x00020000
	// The window has a horizontal scroll bar.
	WS_HSCROLL = 0x00100000
	// The window is initially minimized. Same as the WS_MINIMIZE style.
	WS_ICONIC = 0x20000000
	// The window is initially maximized.
	WS_MAXIMIZE = 0x01000000
	// The window has a maximize button. Cannot be combined with the WS_EX_CONTEXTHELP style. The WS_SYSMENU style must also be specified.
	WS_MAXIMIZEBOX = 0x00010000
	// The window is initially minimized. Same as the WS_ICONIC style.
	WS_MINIMIZE = 0x20000000
	// The window has a minimize button. Cannot be combined with the WS_EX_CONTEXTHELP style. The WS_SYSMENU style must also be specified.
	WS_MINIMIZEBOX = 0x00020000
	// The window is an overlapped window. An overlapped window has a title bar and a border. Same as the WS_TILED style.
	WS_OVERLAPPED = 0x00000000
	// The window is an overlapped window. Same as the WS_TILEDWINDOW style.
	WS_OVERLAPPEDWINDOW = WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU | WS_THICKFRAME | WS_MINIMIZEBOX | WS_MAXIMIZEBOX
	// The window is a pop-up window. This style cannot be used with the WS_CHILD style.
	WS_POPUP = 0x80000000
	// The window is a pop-up window. The WS_CAPTION and WS_POPUPWINDOW styles must be combined to make the window menu visible.
	WS_POPUPWINDOW = WS_POPUP | WS_BORDER | WS_SYSMENU
	// The window has a sizing border. Same as the WS_THICKFRAME style.
	WS_SIZEBOX = 0x00040000
	// The window has a window menu on its title bar. The WS_CAPTION style must also be specified.
	WS_SYSMENU = 0x00080000
	// The window is a control that can receive the keyboard focus when the user presses the TAB key. Pressing the TAB key changes the keyboard focus to the next control with the WS_TABSTOP style. You can turn this style on and off to change dialog box navigation. To change this style after a window has been created, use the SetWindowLong function. For user-created windows and modeless dialogs to work with tab stops, alter the message loop to call the IsDialogMessage function.
	WS_TABSTOP = 0x00010000
	// The window has a sizing border. Same as the WS_SIZEBOX style.
	WS_THICKFRAME = 0x00040000
	// The window is an overlapped window. An overlapped window has a title bar and a border. Same as the WS_OVERLAPPED style.
	WS_TILED = 0x00000000
	// The window is an overlapped window. Same as the WS_OVERLAPPEDWINDOW style.
	WS_TILEDWINDOW = WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU | WS_THICKFRAME | WS_MINIMIZEBOX | WS_MAXIMIZEBOX
	// The window is initially visible. This style can be turned on and off by using the ShowWindow or SetWindowPos function.
	WS_VISIBLE = 0x10000000
	// The window has a vertical scroll bar.
	WS_VSCROLL = 0x00200000
)
