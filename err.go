package my

import (
	"errors"
	"strings"
)

func panicIf(err error) {
	if err != nil {
		dumpAt(2, err)
		//log.Fatal(err)
		panic(err)
	}
}
func PanicIf(err error) {
	panicIf(err)
}
func Must(err error) {
	panicIf(err)
}

type Error struct {
	error
	trace Trace
}
func (Error) New(text string) Error {
	return Error{
		error: errors.New(text),
		trace: Trace{}.New().SkipFile(1).Local(),
	}
}
func (error Error) Error() string {
	trace := make([]string, 0, len(error.trace))
	for _, frame := range error.trace { trace = append(trace, "- " + frame.String()) }
	return error.error.Error() + "\nTrace:\n" + strings.Join(trace, "\n")
}
