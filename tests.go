package my

import (
	"reflect"
	"testing"
)

func areEqual(a interface{}, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}
func fail(t *testing.T, context ...interface{}) {
	if t == nil {
		for _, arg := range context { Dump(arg) }
		panic(context)
	}
	t.Error(context...)
}
func assert(t *testing.T, statement bool, context ...interface{}) {
	context = append([]interface{}{"assertion failed"}, context...)
	if !statement { fail(t, context...) }
}
func assertEquals(t *testing.T, a interface{}, b interface{}, context ...interface{}) {
	args := append([]interface{}{"vars are not equal", a, b}, context...)
	if !areEqual(a, b) { fail(t, args...) }
}
func assertNotEqual(t *testing.T, a interface{}, b interface{}) {
	if areEqual(a, b) { fail(t, "vars are equal", a, b) }
}
