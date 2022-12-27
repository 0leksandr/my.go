package my

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"math"
	"reflect"
	"sync"
	"testing"
)

func areEqual(a, b interface{}, opts ...cmp.Option) bool {
	//return reflect.DeepEqual(a, b)

	var localTypesInterfaces []interface{}
	var isStruct func(reflect.Type) bool
	isStruct = func(_type reflect.Type) bool {
		switch _type.Kind() {
			case reflect.Struct: return true
			case reflect.Ptr:    return isStruct(_type.Elem())
			default:             return false
		}
	}
	for _, _type := range getTypes(pkgName(Trace{}.New().SkipFile(1)[0].File) + ".") {
		if isStruct(_type) {
			localTypesInterfaces = append(localTypesInterfaces, reflect.Zero(_type).Interface())
		}
	}
	//goland:noinspection GoVetCopyLock
	opts = append(
		opts,
		cmp.Comparer(func(a, b error) bool { return a.Error() == b.Error() }),
		cmp.Comparer(func(_, _ sync.Mutex) bool { return true }),
		cmp.AllowUnexported(localTypesInterfaces...),
	)
	return cmp.Equal(a, b, opts...)
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
	if !areEqual(a, b) { Fail(t, args...) }
}
func AssertNotEqual(t *testing.T, a interface{}, b interface{}) {
	if areEqual(a, b) { Fail(t, "vars are equal", a, b) }
}
func ApproxEqual(a, b interface{}) bool {
	return areEqual(
		a,
		b,
		cmp.Comparer(func(a, b float32) bool { return floatsEqual(float64(a), float64(b), 1e-7) }),
		cmp.Comparer(func(a, b float64) bool { return floatsEqual(a, b, 1e-14) }),
	)
}

func floatsEqual(a, b, epsilon float64) bool {
	if b == 0 { return a == 0 }
	return math.Abs(1. - a / b) < epsilon
}
