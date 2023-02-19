package main

import (
	"testing"

	"github.com/operdies/windows-nats-shell/pkg/gfx/colors"
)

func TestColorParser(t *testing.T) {
	cases := map[string][4]float32{
		"#123":   {0, 0x11, 0x22, 0x33},
		"123":    {0, 0x11, 0x22, 0x33},
		"12":     {-1},
		"#abcd":  {0xaa, 0xbb, 0xcc, 0xdd},
		"abcd":   {0xaa, 0xbb, 0xcc, 0xdd},
		"abcdef": {0x00, 0xab, 0xcd, 0xef},
		"qrz":    {-1},
	}

	for s, r := range cases {
		actual, error := colors.StringToColor(s)
		if r[0] == -1 {
			if error == nil {
				t.Fatalf("Expected error.")
			}
			continue
		}
		expected := [4]float32{r[1] / 255, r[2] / 255, r[3] / 255, r[0] / 255}
		if expected != actual {
			t.Fatalf("Expected %v, got %v", expected, actual)
		}
	}
}
