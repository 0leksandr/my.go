package my

import (
	"context"
	"sync"
	"time"
)

func Reservoir[T any](in <-chan T, size int) <-chan T {
	out := make(chan T)
	reservoir := make([]T, 0, size)
	var draining Locker
	draining.Lock()
	var mx sync.Mutex // accessing reservoir or draining state
	inClosed := false
	go func() {
		for value := range in {
			mx.Lock()
			reservoir = append(reservoir, value)
			if len(reservoir) == size {
				// PRIORITY: use Archive as a fallback
			}
			draining.Unlock()
			mx.Unlock()
		}
		mx.Lock()
		inClosed = true
		draining.Unlock()
		mx.Unlock()
	}()
	go func() {
		for {
			draining.Wait()
			for {
				mx.Lock()
				if len(reservoir) > 0 {
					value := reservoir[0]
					reservoir = reservoir[1:]
					mx.Unlock()
					out <- value
				} else {
					if inClosed {
						close(out)
						return
					}
					draining.Lock()
					mx.Unlock()
					break
				}
			}
		}
	}()

	return out
}
func Delayer[T any](channel <-chan T, min, max time.Duration, operation func([]T)) {
	var cancelMin, cancelMax context.CancelFunc
	var mutex sync.Mutex
	var values []T
	doOperation := func() {
		mutex.Lock()
		if cancelMax != nil {
			cancelMax()
			cancelMax = nil
		}
		if cancelMin != nil {
			cancelMin()
			cancelMin = nil
		}
		currentValues := values
		values = nil
		mutex.Unlock()
		operation(currentValues)
	}
	Dispenser(channel, func(batch []T) {
		mutex.Lock()
		values = append(values, batch...)
		if cancelMax == nil { cancelMax = CancellableTimer(max, doOperation) }
		if cancelMin != nil { cancelMin() }
		cancelMin = CancellableTimer(min, doOperation)
		mutex.Unlock()
	})
	doOperation()
}
func Dispenser[T any](channel <-chan T, f func([]T)) {
	for value := range channel {
		values := []T{value}
	    readingResidualValues: for {
			select {
				case nextValue, ok := <- channel:
					if ok {
						values = append(values, nextValue)
					} else {
						break readingResidualValues
					}
				default:
					break readingResidualValues
			}
		}
		f(values)
	}
}
