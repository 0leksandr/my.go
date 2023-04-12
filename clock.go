package my

import (
	"context"
	"io"
	"sort"
	"sync"
	"time"
)

type TimerInterface interface {
	Stop() bool
}

type Clock interface {
	Now() time.Time
	Sleep(time.Duration)
	AfterFunc(time.Duration, func()) TimerInterface
}

var clock Clock = RealClock{}

type RealClock struct {}
func (RealClock) Now() time.Time {
	return time.Now()
}
func (RealClock) Sleep(duration time.Duration) {
	time.Sleep(duration)
}
func (RealClock) AfterFunc(duration time.Duration, f func()) TimerInterface {
	return time.AfterFunc(duration, f)
}

type mockTimerCallee interface { // TODO: rename
	io.Closer
	Call(time.Time)
	Stop()
}
type funcMockTimerCallee struct {
	f func()
}
func (*funcMockTimerCallee) New(f func()) *funcMockTimerCallee {
	return &funcMockTimerCallee{f}
}
func (callee *funcMockTimerCallee) Close() error {
	return nil
}
func (callee *funcMockTimerCallee) Call(time.Time) {
	callee.f()
	callee.Stop()
}
func (callee *funcMockTimerCallee) Stop() {
	callee.f = func() {}
}

type mockTimer struct {
	start     time.Time
	original  time.Duration
	remaining time.Duration
	callee    mockTimerCallee
}
func (*mockTimer) New(duration time.Duration, callee mockTimerCallee) *mockTimer {
	return &mockTimer{
		clock.Now(),
		duration,
		duration,
		callee,
	}
}
func (t *mockTimer) Close() error {
	return t.callee.Close()
}
func (t *mockTimer) Done() bool {
	return t.remaining <= 0
}
func (t *mockTimer) Remaining() time.Duration {
	return t.remaining
}
func (t *mockTimer) Passed(duration time.Duration) {
	t.remaining -= duration
	if t.remaining <= 0 {
		t.callee.Call(t.start.Add(t.original))
	}
}
func (t *mockTimer) Stop() bool {
	t.callee.Stop()
	return true
}

type MockClock struct {
	now    time.Time
	timers []*mockTimer
	mutex  sync.Mutex
}
func (*MockClock) New() *MockClock {
	return &MockClock{now: time.Now()}
}
func (clock *MockClock) Close() error {
	for _, timer := range clock.timers { Must(timer.Close()) }
	return nil
}
func (clock *MockClock) Now() time.Time {
	return clock.now
}
func (clock *MockClock) Sleep(duration time.Duration) {
	clock.mutex.Lock()
	clock.passed(duration)
	clock.mutex.Unlock()
}
func (clock *MockClock) AfterFunc(duration time.Duration, f func()) TimerInterface {
	return clock.addTimer(duration, (*funcMockTimerCallee)(nil).New(f))
}
func (clock *MockClock) passed(duration time.Duration) {
	if len(clock.timers) == 0 {
		clock.now = clock.now.Add(duration)
	} else {
		passed := func(duration time.Duration) {
			clock.now = clock.now.Add(duration)
			for _, timer := range clock.timers { timer.Passed(duration) }
		}
		if duration >= clock.timers[0].Remaining() {
			passedNow := clock.timers[0].Remaining()
			passed(passedNow)
			for len(clock.timers) > 0 {
				timer := clock.timers[0]
				if timer.Done() {
					Must(timer.Close())
					clock.timers = clock.timers[1:]
				} else {
					break
				}
			}
			duration -= passedNow
			if duration > 0 { clock.passed(duration) }
		} else {
			passed(duration)
		}
	}
}
func (clock *MockClock) addTimer(duration time.Duration, callee mockTimerCallee) *mockTimer {
	clock.mutex.Lock()
	defer clock.mutex.Unlock()

	timer := (*mockTimer)(nil).New(duration, callee)
	position := sort.Search(len(clock.timers), func(i int) bool {
		return clock.timers[i].Remaining() > duration
	})
	clock.timers = InsertAt(clock.timers, position, timer)

	return timer
}

func MockTime(f func(*MockClock)) {
	prevClock := clock

	mockClock := (*MockClock)(nil).New()
	clock = mockClock
	f(mockClock)
	Must(mockClock.Close())

	clock = prevClock
}

func CancellableTimer(timeout time.Duration, callback func()) context.CancelFunc {
	timer := clock.AfterFunc(timeout, callback)
	return func() { timer.Stop() }
}
