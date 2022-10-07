package colors

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	Purple = "00ac21c4"
	Red    = "f00"
	Green  = "0f0"
	Blue   = "00f"
)

func StringToColor(s string) (color [4]float32, err error) {
	// strip leading pound sign
	original := s
	s = strings.TrimSpace(s)
	if s[0] == '#' {
		s = s[1:]
	}
	// interpret e.g. #123 as #112233
	if len(s) == 3 || len(s) == 4 {
		s2 := ""
		for i := range s {
			s2 += string([]byte{s[i], s[i]})
		}

		s = string(s2)
	}
	if len(s) == 6 {
		s = "00" + s
	}
	if len(s) != 8 {
		err = fmt.Errorf("Invalid input: %v. Input should be of the from: #argb or #aarrggbb. '#' and 'a' are optional", original)
		return
	}
	col := func(c string) (v float32, err error) {
		v2, err := strconv.ParseUint(c, 16, 8)
		if err != nil {
			return
		}
		v = float32(v2) / 255
		return
	}
	color = [4]float32{}
	for i := 0; i < 4; i++ {
		start := i * 2
		end := i*2 + 2
		c, err := col(s[start:end])
		if err != nil {
			return color, err
		}
		color[i] = c
	}

	// Put alpha last
	return [4]float32{color[1], color[2], color[3], color[0]}, nil
}
