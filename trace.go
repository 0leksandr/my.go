package my

import (
	"fmt"
	"regexp"
	"runtime"
)

type Frame struct {
	File string
	Line int
}
func (frame Frame) String() string {
	return fmt.Sprintf("%v:%d", frame.File, frame.Line)
}

type Frames []Frame
func (frames Frames) SkipFile(n int) Frame {
	file := frames[0].File
	i := 0
	for n > 0 {
		if frames[i].File == file {
			i++
		} else {
			n--
			file = frames[i].File
		}
	}
	return frames[i]
}

func Trace(full bool) Frames {
	size := 1<<8
	pc := make([]uintptr, size, size)
	runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc)
	var trace Frames
	add := func(frame runtime.Frame) {
		trace = append(trace, Frame{
			File: frame.File,
			Line: frame.Line,
		})
	}
	frame, more := frames.Next()
	projectRoot := regexp.MustCompile("^(.+/)[^/]+$").FindStringSubmatch(frame.File)[1]
	add(frame)
	for more {
		frame, more = frames.Next()
		if full || regexp.MustCompile(fmt.Sprintf("^%s[^/]+$", projectRoot)).MatchString(frame.File) {
			add(frame)
		}
	}
	return trace
}
