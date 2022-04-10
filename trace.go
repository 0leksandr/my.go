package my

import (
	"fmt"
	"regexp"
	"runtime"
)

func Trace(full bool) []string {
	size := 1<<8
	pc := make([]uintptr, size, size)
	runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc)
	var trace []string
	projectRoot := ""
	for {
		frame, more := frames.Next()
		add := true
		if projectRoot == "" {
			projectRoot = regexp.MustCompile("^(.+/)[^/]+$").FindStringSubmatch(frame.File)[1]
		} else {
			if !full {
				add = regexp.MustCompile(fmt.Sprintf("^%s[^/]+$", projectRoot)).MatchString(frame.File)
			}
		}
		if add { trace = append(trace, fmt.Sprintf("%v:%d", frame.File, frame.Line)) }
		if !more { break }
	}
	return trace
}
