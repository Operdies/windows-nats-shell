package utils

import (
	"time"

	"gopkg.in/yaml.v3"
)

func EncodeAny[T any](value T) []byte {
	result, _ := yaml.Marshal(value)
	return result
}

func DecodeAny[T any](buffer []byte) T {
	var result T
	yaml.Unmarshal(buffer, &result)
	return result
}

// Return a function which will batch calls that happen within a given interval
// GO does not guarantee that the first responding channel selected.
// The only guarantee is that the function will eventually be called
// The given function will be called unless a different call is made to
// the returned function within `minInterval`. If successive calls are made
// within `minInterval`, then the function is called after `maxInterval`.
// Additional calls will start a new batch.
func Batcher(f func(), minInterval time.Duration, maxInterval time.Duration) func() {
	addToBatch := make(chan bool)
	var minTimer time.Timer
	var maxTimer time.Timer

	handleBatch := func() {
		for {
			select {
			case <-addToBatch:
				// Extend the timer
				minTimer = *time.NewTimer(minInterval)
			case <-minTimer.C:
				return
			case <-maxTimer.C:
				return
			}
		}
	}

	go func() {
		for <-addToBatch {
			minTimer = *time.NewTimer(minInterval)
			maxTimer = *time.NewTimer(maxInterval)
			handleBatch()
			f()
		}
	}()

	return func() {
		addToBatch <- true
	}
}
