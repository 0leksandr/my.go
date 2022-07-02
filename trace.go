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

type Trace []Frame
func (Trace) New() Trace {
	size := 1<<8
	pc := make([]uintptr, size)
	runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc)
	var trace Trace
	for {
		frame, more := frames.Next()
		trace = append(trace, Frame{
			File: frame.File,
			Line: frame.Line,
		})
		if !more { break }
	}

	return trace
}
func (trace Trace) SkipFile(n int) Trace {
	file := trace[0].File
	i := 0
	for n > 0 {
		if trace[i].File == file {
			i++
		} else {
			n--
			file = trace[i].File
		}
	}
	return trace[i:]
}
func (trace Trace) Local() Trace {
	var localTrace Trace
	sep := string(os.PathSeparator)
	firstFrame := trace[0]
	projectRoot := regexp.MustCompile(fmt.Sprintf("^(.+%s)[^%s]+$", sep, sep)).FindStringSubmatch(firstFrame.File)[1]
	projectRootRe := regexp.MustCompile(fmt.Sprintf("^%s([^%s]+)$", projectRoot, sep))
	for _, frame := range trace {
		if projectRootRe.MatchString(frame.File) {
			frame.File = projectRootRe.FindStringSubmatch(frame.File)[1]
			localTrace = append(localTrace, frame)
		}
	}

	return localTrace
}

func GetTrace(full bool) Trace {
	trace := Trace{}.New()[1:]
	if !full { trace = trace.Local() }
	return trace
}
