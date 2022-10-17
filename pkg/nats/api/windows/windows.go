package windows

import (
	"math"

	"github.com/operdies/windows-nats-shell/pkg/winapi/wintypes"
)

const (
	// Poll for visible windows
	GetWindows = "Windows.GetWindows"
	// Subscribe to window events
	WindowsUpdated = "Windows.Updated"
	// Poll for window focused state
	IsWindowFocused = "Window.Focused"
	// Attempt to bring the selected window to the foreground
	FocusWindow = "Window.SetFocus"
	// MoveWindow the window
	MoveWindow = "Window.Move"
	// ResizeWindow the window
	ResizeWindow = "Window.Resize"
	// MinimizeWindow the window
	MinimizeWindow = "Window.Minimize"
	// RestoreWindow the window
	RestoreWindow = "Window.Restore"
	// MaximizeWindow the window
	MaximizeWindow = "Window.Maximize"
	// Focus the previous window
	FocusPrevious = "Window.FocusPrevious"
	// Focus the next window
	FocusNext = "Window.FocusNext"
	// Hide window border
	HideBorder = "Window.HideBorder"
	// Show window border
	ShowBorder = "Window.ShowBorder"
	// Toggle window border
	ToggleBorder = "Window.ToggleBorder"
)

type WindowCardinals = int

const (
	Top         WindowCardinals = 0b0001
	Left                        = 0b0010
	Bottom                      = 0b0100
	Right                       = 0b1000
	TopLeft                     = Top | Left
	TopRight                    = Top | Right
	BottomLeft                  = Bottom | Left
	BottomRight                 = Bottom | Right
)

func dist(p1 wintypes.POINT, p2 wintypes.POINT) float64 {
	x1 := float64(p1.X)
	x2 := float64(p2.X)
	y1 := float64(p1.Y)
	y2 := float64(p2.Y)
	return math.Sqrt(math.Pow(x1-x2, 2) + math.Pow(y1-y2, 2))
}

// Given a cardinal direction, get a point representing that direction on the rectangle
func GetPoints(r wintypes.RECT) map[WindowCardinals]wintypes.POINT {
	c := func(a, b int32) int32 { return a + (b-a)/2 }
	p := func(x, y int32) wintypes.POINT { return wintypes.POINT{X: wintypes.LONG(x), Y: wintypes.LONG(y)} }
	return map[WindowCardinals]wintypes.POINT{
		Top:         p(c(r.Left, r.Right), r.Top),
		Left:        p(r.Left, c(r.Top, r.Bottom)),
		Bottom:      p(c(r.Left, r.Right), r.Bottom),
		Right:       p(r.Right, c(r.Top, r.Bottom)),
		TopLeft:     p(r.Left, r.Top),
		TopRight:    p(r.Right, r.Top),
		BottomLeft:  p(r.Left, r.Bottom),
		BottomRight: p(r.Right, r.Bottom),
	}
}

func GetNearestCardinal(p wintypes.POINT, r wintypes.RECT) (result WindowCardinals) {
	candidates := GetPoints(r)
	closest := math.Inf(1)
	for k, v := range candidates {
		d := dist(p, v)
		if d < closest {
			result = k
			closest = d
		}
	}
	return
}

func cardinalToString(c WindowCardinals) string {
	var verb string
	switch c {
	case Top:
		verb = "Top"
	case Left:
		verb = "Left"
	case Bottom:
		verb = "Bottom"
	case Right:
		verb = "Right"
	case TopLeft:
		verb = "TopLeft"
	case TopRight:
		verb = "TopRight"
	case BottomLeft:
		verb = "BottomLeft"
	case BottomRight:
		verb = "BottomRight"
	}
	return verb
}
