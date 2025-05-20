package my

import (
	"fmt"
	"runtime"
)

func fileLine(skip int) string {
	_, file, line, _ := runtime.Caller(skip + 1)
	return fmt.Sprintf("%v:%d", file, line)
}
func dumpAt(skip int, values ...any) {
	fl := fileLine(skip + 1)
	if len(values) == 0 { fmt.Printf("%v\n", fl) }
	for _, val := range values {
		fmt.Printf("%v %#v\n", fl, val)
		if err, ok := val.(error); ok { fmt.Printf("%v %s\n", fl, err.Error()) }
	}
}
func Dump(values ...any) {
	dumpAt(1, values...)
}

func Dump3(values ...any) {
	fl := fileLine(1)
	if len(values) == 0 { fmt.Println(fl) }
	for _, val := range values {
		fmt.Println(fl + " " + formatter{}.New().Format(val))
	}
}
