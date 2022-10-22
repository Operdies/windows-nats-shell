package windows

import (
	"math"
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

type CardinalDirection = int

const (
	Top CardinalDirection = 1 << iota
	Left
	Bottom
	Right
	TopLeft     = Top | Left
	TopRight    = Top | Right
	BottomLeft  = Bottom | Left
	BottomRight = Bottom | Right
)

// Given a cardinal direction, get a point representing that direction on the rectangle
func GetPoints(r Rect) map[CardinalDirection]Point {
	c := func(a, b int32) int32 { return a + (b-a)/2 }
	p := func(x, y int32) Point { return Point{X: x, Y: y} }
	return map[CardinalDirection]Point{
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

func GetNearestCardinal(p Point, r Rect) (result CardinalDirection) {
	candidates := GetPoints(r)
	closest := math.Inf(1)
	for k, v := range candidates {
		d := p.DistanceTo(v)
		if d < closest {
			result = k
			closest = d
		}
	}
	return
}

func cardinalToString(c CardinalDirection) string {
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

type Rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

func (r Rect) Translate(x, y int) Rect {
	return Rect{
		Left:   r.Left + int32(x),
		Right:  r.Right + int32(x),
		Top:    r.Top + int32(y),
		Bottom: r.Bottom + int32(y),
	}
}

func (r Rect) Width() int32 {
	return r.Right - r.Left
}

func (r Rect) Height() int32 {
	return r.Bottom - r.Top
}

func (r Rect) CenterAround(other Point) Rect {
	rect := Rect{
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
func (r Rect) Animate(final Rect, frames int, animateSize bool) []Rect {
	to := final
	if !animateSize {
		to = r.CenterIn(to)
	}
	lefts := animate(float64(r.Left), float64(to.Left), frames)
	rights := animate(float64(r.Right), float64(to.Right), frames)
	tops := animate(float64(r.Top), float64(to.Top), frames)
	bottoms := animate(float64(r.Bottom), float64(to.Bottom), frames)
	result := make([]Rect, 0, frames)
	for i := 0; i < frames; i++ {
		result = append(result, Rect{
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
func (r Rect) GetPointOnPerimeterLineMethod(point float64) Point {
	// Rotate the point 1/8th around the perimeter so '0' becomes the center of the top line
	point += 0.125
	// put point in the range 0 <= point < 1
	point = point - math.Floor(point)
	// get new value in range 0 <= phase < 4
	phase := point * 4
	var X, Y int32
	var xMod, yMod float64
	if phase < 1 {
		// Top side
		xMod = phase
		yMod = 0
	} else if phase < 2 {
		// Right side
		xMod = 1
		yMod = phase - 1
	} else if phase < 3 {
		// Bottom side
		yMod = 1
		xMod = 3 - phase
	} else {
		// Left side
		xMod = 0
		yMod = 4 - phase
	}
	X = int32(r.Left) + int32(float64(r.Width())*xMod)
	Y = int32(r.Top) + int32(float64(r.Height())*yMod)

	return Point{X, Y}
}
func (r Rect) GetPointOnPerimeterCircleMethod(point float64) Point {
	// put point in the range 0 <= point < 1
	point = point - math.Floor(point)
	rad := point * math.Pi * 2
	sin := math.Sin(rad)
	cos := math.Cos(rad)

	// Extrapolate the point on the circle onto the closest side of the surrounding square
	factor := 1.0 / math.Max(math.Abs(cos), math.Abs(sin))
	cos *= factor
	sin *= factor

	return Point{
		X: int32(r.Left) + (int32(r.Width()/2) + int32(math.Round(sin*float64(r.Width()/2)))),
		Y: int32(r.Top) + (int32(r.Height()/2) - int32(math.Round(cos*float64(r.Height()/2)))),
	}
}

func (r Rect) CenterIn(other Rect) Rect {
	otherHalfWidth := other.Width() / 2
	rHalfWidth := r.Width() / 2
	otherHalfHeight := other.Height() / 2
	rHalfHeight := r.Height() / 2
	return Rect{
		Left:   other.Left + (otherHalfWidth - rHalfWidth),
		Right:  other.Left + (otherHalfWidth + rHalfWidth),
		Top:    other.Top + (otherHalfHeight - rHalfHeight),
		Bottom: other.Top + (otherHalfHeight + rHalfHeight),
	}
}

// Align `r` to `o` along one or two axes
func (r Rect) Align(o Rect, c CardinalDirection) Rect {
	r = r.CenterIn(o)
	directions := []CardinalDirection{Top, Left, Bottom, Right}
	heightDelta := r.Height() - o.Height()
	widthDelta := r.Width() - o.Width()
	for _, v := range directions {
		if c&v > 0 {
			if v == Top {
				r = r.Translate(0, int(heightDelta)/2)
			} else if v == Bottom {
				r = r.Translate(0, -int(heightDelta)/2)
			} else if v == Left {
				r = r.Translate(int(widthDelta)/2, 0)
			} else if v == Right {
				r = r.Translate(-int(widthDelta)/2, 0)
			}
		}
	}
	return r
}

func (r Rect) ScaleY(factor float64) Rect {
	centered := r.CenterIn(Rect{0, 0, 0, 0})
	centered.Top = int32(factor * float64(centered.Top))
	centered.Bottom = int32(factor * float64(centered.Bottom))
	return centered.CenterIn(r)
}

func (r Rect) ScaleX(factor float64) Rect {
	centered := r.CenterIn(Rect{0, 0, 0, 0})
	centered.Left = int32(factor * float64(centered.Left))
	centered.Right = int32(factor * float64(centered.Right))
	return centered.CenterIn(r)
}
func (r Rect) Scale(factor float64) Rect {
	centered := r.CenterIn(Rect{0, 0, 0, 0})
	centered.Left = int32(factor * float64(centered.Left))
	centered.Right = int32(factor * float64(centered.Right))
	centered.Top = int32(factor * float64(centered.Top))
	centered.Bottom = int32(factor * float64(centered.Bottom))
	return centered.CenterIn(r)
}

func (r Rect) PadY(y int32) Rect {
	return r.Pad(0, y)
}
func (r Rect) PadX(x int32) Rect {
	return r.Pad(x, 0)
}

func (r Rect) Pad(x, y int32) Rect {
	return Rect{
		Left:   r.Left + x,
		Right:  r.Right - x,
		Top:    r.Top + y,
		Bottom: r.Bottom - y,
	}
}

type Point struct {
	X, Y int32
}

func (p1 Point) DistanceTo(p2 Point) float64 {
	x1 := float64(p1.X)
	x2 := float64(p2.X)
	y1 := float64(p1.Y)
	y2 := float64(p2.Y)
	return math.Sqrt(math.Pow(x1-x2, 2) + math.Pow(y1-y2, 2))
}
func (p Point) Add(p2 Point) Point {
	return Point{X: p.X + p2.X, Y: p.Y + p2.Y}
}
func (p Point) Sub(p2 Point) Point {
	return Point{X: p.X - p2.X, Y: p.Y - p2.Y}
}
