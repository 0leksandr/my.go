package my

import (
	"sync"
)

type Locker struct {
	wg     sync.WaitGroup
	mutex  sync.Mutex
	locked bool
}
func (locker *Locker) Lock() {
	locker.mutex.Lock()
	if !locker.locked {
		locker.locked = true
		locker.wg.Add(1)
	}
	locker.mutex.Unlock()
}
func (locker *Locker) Unlock() {
	locker.mutex.Lock()
	if locker.locked {
		locker.locked = false
		locker.wg.Done()
	}
	locker.mutex.Unlock()
}
func (locker *Locker) Wait() {
	locker.wg.Wait()
}
