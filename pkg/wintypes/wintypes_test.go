package wintypes

import (
	"testing"
)

func TestRect(t *testing.T) {
	{
		r := RECT{100, 100, 300, 200}
		r2 := r.Scale(1.5)

		if r.Height() != 100 {
			t.Errorf("Incorrect height")
		}
		if r.Width() != 200 {
			t.Errorf("Incorrect width")
		}

		if r2.Height() != 150 {
			t.Errorf("Height scaling failed")
		}
		if r2.Width() != 300 {
			t.Errorf("Width scaling failed")
		}

		t.Logf("%v\n", r)
		t.Logf("%v\n", r2)
		t.Logf("%v\n", r2.CenterIn(r))
	}

	{
		r := RECT{0, 0, 10, 10}
		r2 := r.CenterAround(POINT{5, 5})

		if r.Left != r2.Left || r.Right != r2.Right || r.Bottom != r2.Bottom || r.Top != r2.Top {
			t.Errorf("Centering failed")
		}

		r = r.CenterAround(POINT{0, 0})
		if r.Left != -5 || r.Right != 5 || r.Top != -5 || r.Bottom != 5 {
			t.Error("Centering failed!")
		}
	}
}

func TestRectPerimeter(t *testing.T) {
	r := RECT{100, 100, 200, 200}
	expected := map[float64]POINT{
		0:     {150, 100},
		0.125: {200, 100},
		0.25:  {200, 150},
		0.375: {200, 200},
		0.5:   {150, 200},
		0.625: {100, 200},
		0.75:  {100, 150},
		0.875: {100, 100},
		1.0:   {150, 100},
	}
	for f, p := range expected {
		actual := r.GetPointOnPerimeter(f)
		if actual.X != p.X || actual.Y != p.Y {
			t.Errorf("Got: %v, expected: %v", actual, p)
		}
	}
	// f := 0.0
	// for f < 1.1 {
	// 	p := r.GetPointOnPerimeter(f)
	// 	fmt.Printf("f: %v, p: %+v\n", f, p)
	// 	f += 0.125
	// }
}
