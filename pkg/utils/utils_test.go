package utils

import (
	"testing"
	"time"
)

func TestBatching(t *testing.T) {
	calls := 0
	b := Batcher(func() { calls += 1 }, time.Millisecond, time.Second)

	time.Sleep(time.Millisecond)
	for i := 0; i < 30; i++ {
		for j := 0; j < 500; j++ {
			b()
		}

		if calls != i {
			t.Fatalf("We should still be waiting for batch %d.", i+1)
		}
		time.Sleep(time.Millisecond * 20)
		if calls != i+1 {
			t.Fatalf("Expected %d batched calls. Got %d.", i+1, calls)
		}
	}
}
