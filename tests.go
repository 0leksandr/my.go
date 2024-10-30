package my

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"math"
	"path"
	"reflect"
	"sync"
	"testing"
	"time"
)

func areEqual(a, b any, opts ...cmp.Option) bool {
	//return reflect.DeepEqual(a, b)

	var localTypesInterfaces []any
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
		cmp.Comparer(func(a, b time.Time) bool { return a.Round(time.Duration(0)) == b.Round(time.Duration(0)) }),
		cmp.AllowUnexported(localTypesInterfaces...),
	)
	return cmp.Equal(a, b, opts...)
}
func Fail(t *testing.T, context ...any) {
	fmt.Println(Trace{}.New().SkipFile(1)[0])
	for _, c := range context { Dump2(c) }
	Dump2(Trace{}.New().SkipFile(1).Local())
	if t != nil {
		t.Fail()
	} else {
		panic("check failed")
	}
}
func Assert(t *testing.T, statement bool, context ...any) {
	context = append([]any{"assertion failed"}, context...)
	if !statement { Fail(t, context...) }
}
func AssertEquals(t *testing.T, a, b any, context ...any) {
	if !areEqual(a, b) {
		Fail(t, append([]any{"vars are not equal", a, b}, context...)...)
	}
}
func AssertNotEqual(t *testing.T, a, b any) {
	if areEqual(a, b) { Fail(t, "vars are equal", a, b) }
}
func ApproxEqual(a, b any) bool {
	return areEqual(
		a,
		b,
		cmp.Comparer(func(a, b float32) bool { return floatsEqual(float64(a), float64(b), 1e-7) }),
		cmp.Comparer(func(a, b float64) bool { return floatsEqual(a, b, 1e-14) }),
	)
}
func AssertNil(t *testing.T, value any) {
	Assert(t, isNil(value), "value is not nil", value)
}

func TestTypes(t *testing.T) {
	testTypes(t, nil)
}
func testTypes(t *testing.T, ignored []string) {
	parsedPackage := parseTypes(path.Dir(GetTrace(true).SkipFile(1)[0].File))
	types := Types(true)
	for structName, parsedStruct := range parsedPackage.structs {
		if !InArray(structName, ignored) {
			for _, embeddedName := range parsedStruct.embedded {
				if embeddedStruct, isLocalStruct := parsedPackage.structs[embeddedName.KeyName()]; isLocalStruct {
					Assert(
						t,
						parsedStruct.Overrides(embeddedStruct),
						"struct does not override", structName, embeddedName,
					)
				} else if embeddedInterface, isLocalInterface := parsedPackage.interfaces[embeddedName.KeyName()]
					isLocalInterface {
					Assert(
						t,
						parsedStruct.Implements(embeddedInterface),
						"struct does not implement", structName, embeddedName,
					)
				} else {
					embeddedReal := ArrayFilter(types, func(_type reflect.Type) bool {
						return embeddedName.EqualsReal(_type)
					})
					if len(embeddedReal) != 1 {
						panic(fmt.Sprintf(
							"embedded type not found: %s %s(%d)",
							structName,
							embeddedName,
							len(embeddedReal),
						))
					}
					Assert(
						t,
						parsedStruct.ImplementsReal(embeddedReal[0]),
						"struct does not implement", structName, embeddedName,
					)
				}
			}
		}
	}
}

func floatsEqual(a, b, epsilon float64) bool {
	if b == 0 { return a == 0 }
	return math.Abs(1. - a / b) < epsilon
}
func isNil(value any) bool {
	if value == nil { return true }
	switch reflect.TypeOf(value).Kind() {
		case reflect.Ptr,
			reflect.Map,
			reflect.Array,
			reflect.Chan,
			reflect.Slice:
				return reflect.ValueOf(value).IsNil()
		default:
			return false
	}
}
