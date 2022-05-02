package my

import (
	"fmt"
	"os"
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
	if !more { panic("weird frames") }
	var projectRootRe *regexp.Regexp
	if !full {
		ps := string(os.PathSeparator)
		projectRoot := regexp.MustCompile(fmt.Sprintf("^(.+%s)[^%s]+$", ps, ps)).FindStringSubmatch(frame.File)[1]
		projectRootRe = regexp.MustCompile(fmt.Sprintf("^%s[^%s]+$", projectRoot, ps))
	}
	add(frame)
	for more {
		frame, more = frames.Next()
		if projectRootRe == nil || projectRootRe.MatchString(frame.File) { add(frame) }
	}
	return trace
}
