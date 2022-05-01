package my

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestDump(t *testing.T) {
	Dump("hi")
}
func TestTrace(t *testing.T) {
	currentDir, err := os.Getwd()
	PanicIf(err)
	assertEquals(t, Trace(false), Frames{{currentDir + "/my_test.go", 16}})

	fullTrace := Trace(true)
	assert(t, len(fullTrace) == 3, fullTrace)
}
func TestSdump2(t *testing.T) {
	assertEquals(
		t,
		Sdump2(struct {
			int          int
			string       string
			sliceFloat   []float32
			mapStringInt map[string]int
		}{
			int:        42,
			string:     "test",
			sliceFloat: []float32{1.2, 3.4},
			mapStringInt: map[string]int{
				"key1": 1,
				"key2": 2,
			},
		}),
		`struct { int int; string string; sliceFloat []float32; mapStringInt map[string]int }{
    int: 42,
    string: (len=4) "test",
    sliceFloat: []float32{
        (float32) 1.2,
        (float32) 3.4
    },
    mapStringInt: map[string]int{
        (string) (len=4) "key1": (int) 1,
        (string) (len=4) "key2": (int) 2
    },
}`,
	)
}
func TestPanicIf(t *testing.T) {
	if false {
		PanicIf(errors.New("test"))
	}
	if false {
		Must(errors.New("test"))
	}
}
func TestRevert(t *testing.T) {
	assertEquals(t, Revert([]int{1, 2, 3, 4}).([]int), []int{4, 3, 2, 1})
}
func TestRemove(t *testing.T) {
	slice := []int{1, 2, 3, 4}
	slice = Remove(slice, 2).([]int)
	assertEquals(t, slice, []int{1, 2, 4})
	slice = Remove(slice, 2).([]int)
	assertEquals(t, slice, []int{1, 2})
	slice = Remove(slice, 0).([]int)
	assertEquals(t, slice, []int{2})
	slice = Remove(slice, 0).([]int)
	assertEquals(t, slice, []int{})
}
func TestInArray(t *testing.T) {
	assert(t, InArray(3, []int{1, 2, 3}))
	assert(t, !InArray("3", []string{"1", "2"}))
	//InArray("1", []int{1, 2, 3})
}
func TestDummyMap(t *testing.T) {
	m := DummyMap{}
	assert(t, !m.Has("test"))
	assert(t, m.Len() == 0)
	if _, ok := m.Get("test"); ok { t.Fail() }
	assertEquals(
		t,
		m.Arr(reflect.TypeOf([]string{})),
		[]string{},
	)

	m.Set("test", "a test")
	assert(t, m.Has("test"))
	assert(t, m.Len() == 1)
	(func() {
		value, ok := m.Get("test")
		assertEquals(t, value, "a test")
		assert(t, ok)
	})()
	assertEquals(
		t,
		m.Arr(reflect.TypeOf([]string{})),
		[]string{"a test"},
	)

	m.Del("test")
	assert(t, !m.Has("test"))
	assert(t, m.Len() == 0)
	if _, ok := m.Get("test"); ok { t.Fail() }
	assertEquals(
		t,
		m.Arr(reflect.TypeOf([]string{})),
		[]string{},
	)
}
func TestTypes(t *testing.T) {
	assertEquals(
		t,
		Types(false),
		[]reflect.Type{
			reflect.TypeOf(Frame{}),
			reflect.TypeOf(Frames{}),
			reflect.TypeOf(ParsedArrayType{}),
			reflect.TypeOf(ParsedFunc{}),
			reflect.TypeOf(ParsedInterface{}),
			reflect.TypeOf(ParsedNamedType{}),
			reflect.TypeOf(ParsedPackage{}),
			reflect.TypeOf(ParsedStruct{}),
			reflect.TypeOf((*ParsedType)(nil)).Elem(),
		},
	)
}
func TestParseTypes(t *testing.T) {
	parsed := ParseTypes()
	testInterface := parsed.Interfaces["TestInterface"]
	testType1 := parsed.Structs["TestType1"]
	testType2 := parsed.Structs["TestType2"]
	assertEquals(
		t,
		testInterface,
		ParsedInterface{
			Methods: map[string]ParsedFunc{
				"TestMethod": {},
			},
		},
	)
	assertEquals(
		t,
		testType1,
		ParsedStruct{
			Methods: map[string]ParsedFunc{},
		},
	)
	assertEquals(
		t,
		testType2,
		ParsedStruct{
			Methods: map[string]ParsedFunc{
				"TestMethod":  {},
				"testMethod2": {Out: []ParsedType{ParsedNamedType{"bool"}}},
			},
		},
	)

	assert(t, !testType1.Implements(testInterface))
	assert(t, testType2.Implements(testInterface))
}

type TestInterface interface { TestMethod() }
//goland:noinspection GoUnusedExportedType
type TestType1 struct { TestInterface }
//goland:noinspection GoUnusedExportedType
type TestType2 struct { TestInterface }
func (t TestType2) TestMethod() {}
func (t TestType2) testMethod2() bool { return true }
