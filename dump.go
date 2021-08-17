package my

import (
	"fmt"
	"runtime"
)

func fileLine(skip int) string {
	_, file, line, _ := runtime.Caller(skip + 1)
	return fmt.Sprintf("%v:%d", file, line)
}
func Dump(values ...interface{}) {
	fl := fileLine(1)
	if len(values) == 0 { fmt.Printf("%v\n", fl) }
	for _, val := range values {
		fmt.Printf("%v %#v\n", fl, val)
		if err, ok := val.(error); ok { fmt.Printf("%v %s\n", fl, err.Error()) }
	}
}
