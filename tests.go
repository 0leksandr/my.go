package my

import (
	"fmt"
	"reflect"
	"testing"
)

func AreEqual(a interface{}, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}
func Fail(t *testing.T, context ...interface{}) {
	fmt.Println(Trace{}.New().SkipFile(1)[0])
	for _, c := range context { Dump2(c) }
	if t != nil {
		t.Fail()
	} else {
		panic("check failed")
	}
}
func Assert(t *testing.T, statement bool, context ...interface{}) {
	context = append([]interface{}{"assertion failed"}, context...)
	if !statement { Fail(t, context...) }
}
func AssertEquals(t *testing.T, a interface{}, b interface{}, context ...interface{}) {
	args := append([]interface{}{"vars are not equal", a, b}, context...)
	if !AreEqual(a, b) { Fail(t, args...) }
}
func AssertNotEqual(t *testing.T, a interface{}, b interface{}) {
	if AreEqual(a, b) { Fail(t, "vars are equal", a, b) }
}
