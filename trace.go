package my

import (
	"fmt"
	"os"
	"path/filepath"
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
	size := 1 << 8
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
	projectRoot := getProjectRoot(trace[0].File)
	projectRootLen := len(projectRoot)
	for _, frame := range trace {
		if len(frame.File) >= projectRootLen && frame.File[:projectRootLen] == projectRoot {
			frame.File = frame.File[projectRootLen:]
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

func getProjectRoot(file string) string {
	sep := string(os.PathSeparator)

	findByTarget := func(target string) string {
		dir := filepath.Dir(file)
		for dir != filepath.Dir(dir) {
			if _, errStat := os.Stat(dir + sep + target); errStat == nil {
				return dir + sep
			}
			dir = filepath.Dir(dir)
		}
		return ""
	}

	if byIdea := findByTarget(".idea"); byIdea != "" {
		return byIdea
	}

	if byGoMod := findByTarget("go.mod"); byGoMod != "" {
		return byGoMod
	}

	if byGit := findByTarget(".git"); byGit != "" {
		return byGit
	}

	return regexp.MustCompile(fmt.Sprintf("^(.+%s)[^%s]+$", sep, sep)).FindStringSubmatch(file)[1]
}
