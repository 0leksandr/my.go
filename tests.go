package my

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"reflect"
	"sync"
	"testing"
)

func AreEqual(a interface{}, b interface{}) bool {
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
	return cmp.Equal(
		a,
		b,
		cmp.Comparer(func(a, b error) bool { return a.Error() == b.Error() }),
		cmp.Comparer(func(_, _ sync.Mutex) bool { return true }),
		cmp.AllowUnexported(localTypesInterfaces...),
	)
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
