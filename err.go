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
func (Error) Wrap(other error) Error {
	if err, ok := other.(Error); ok {
		return err
	}

	return Error{
		error: other,
		trace: Trace{}.New().SkipFile(1).Local(),
	}
}
func (error Error) Error() string {
	trace := make([]string, 0, len(error.trace))
	for _, frame := range error.trace { trace = append(trace, "- " + frame.String()) }
	return error.error.Error() + "\nTrace:\n" + strings.Join(trace, "\n")
}
func (error Error) Unwrap() error {
	return error.error
}
func (error Error) Is(other error) bool {
	if otherError, ok := other.(Error); ok {
		return errors.Is(error.error, otherError.error) &&
			areEqual(error.trace, otherError.trace)
	} else {
		return false
	}
}

func ComboError(error1, error2 error, errors ...error) error {
	errors = append([]error{error1, error2}, errors...)
	errors = ArrayFilter(errors, func(err error) bool { return err != nil })
	switch len(errors) {
		case 0:  return nil
		case 1:  return errors[0]
		default: return Error{}.New(strings.Join(
			ArrayMap(errors, func(err error) string { return err.Error() }),
			"\n---\n",
		))
	}
}
