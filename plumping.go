package my

import "sync"

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
