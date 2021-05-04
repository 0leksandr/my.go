package my

import (
	"fmt"
	"regexp"
	"sync"
	"time"
)

type ProgressBar struct {
	description   string
	total         int64
	current       int64
	width         int
	start         time.Time
	maxWidth      int
	prevNrNotches int
	lastPrint     time.Time
	mutex         *sync.Mutex
}

func (_ ProgressBar) New(description string, totalSteps int64) ProgressBar {
	bar := ProgressBar{
		description: description,
		total:       totalSteps,
		current:     0,
		width:       100,
		start:       time.Now(),
		mutex:       &sync.Mutex{},
	}
	bar.print("", false)

	return bar
}
func (bar *ProgressBar) Add() {
	bar.mutex.Lock()
	bar.current++
	bar.updated()
	bar.mutex.Unlock()
}
func (bar *ProgressBar) UpdateTotal(total int64) {
	bar.total = total
	bar.updated()
}
func (bar ProgressBar) FmtDuration(duration time.Duration) string {
	return regexp.
		MustCompile("^([^.]+\\.\\d{2})\\d*([^0-9]+)$").
		ReplaceAllString(duration.String(), "$1$2")
}
func (bar *ProgressBar) updated() {
	nrNotches := bar.nrNotches()

	if bar.current >= bar.total {
		bar.print(bar.FmtDuration(time.Since(bar.start)), true)
	} else if (nrNotches > bar.prevNrNotches) || (time.Since(bar.lastPrint) > 5 * time.Second) {
		bar.print(
			fmt.Sprintf(
				"%d%% ≈%s",
				bar.current * 100 / bar.total,
				bar.FmtDuration(bar.estimateLeft()),
			),
			false,
		)
	}

	bar.prevNrNotches = nrNotches
}
func (bar *ProgressBar) print(postfix string, newLine bool) {
	max := func(a int, b int) int {
		if a > b {
			return a
		} else {
			return b
		}
	}
	generateText := func(symbol string, nrSymbols int) string {
		_bar := ""
		for c := 0; c < nrSymbols; c++ { _bar += symbol }
		return _bar
	}
	nrNotches := bar.nrNotches()
	notches := generateText("|", nrNotches) + generateText(".", bar.width - nrNotches)
	_bar := bar.description + ": [" + notches + "] " + postfix
	bar.maxWidth = max(bar.maxWidth, len(_bar))
	_bar += generateText(" ", bar.maxWidth - len(_bar))
	if newLine { _bar += "\n" }
	fmt.Print("\r" + _bar)
	bar.lastPrint = time.Now()
}
func (bar ProgressBar) nrNotches() int {
	return int(float64(bar.current) * float64(bar.width) / float64(bar.total))
}
func (bar ProgressBar) estimateLeft() time.Duration {
	return time.Duration(
		float64(time.Since(bar.start).Nanoseconds()) * float64(bar.total - bar.current) / float64(bar.current),
	)
}
func testProgress() {
	for _, nrSteps := range []int64{1, 3, 10, 11, 30, 100, 300, 999, 1000, 1001} {
		sleep := int64(1 * time.Second) / nrSteps
		progress := ProgressBar{}.New(fmt.Sprintf("Test %4d", nrSteps), nrSteps)
		for c := int64(0); c < nrSteps; c++ {
			time.Sleep(time.Duration(sleep))
			progress.Add()
		}
	}
}
