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
	expected := Frames{{currentDir + "/my_test.go", 17}}
	actual := Trace(false)
	if !reflect.DeepEqual(expected, actual) { t.Error(actual) }

	fullTrace := Trace(true)
	if len(fullTrace) != 3 { t.Error(fullTrace) }
}
func TestSdump2(t *testing.T) {
	expected :=
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
}`
	actual := Sdump2(struct {
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
	})
	if actual != expected {
		t.Error(actual)
	}
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
	slice := []int{1, 2, 3, 4}
	expected := []int{4, 3, 2, 1}
	actual := Revert(slice).([]int)
	if !reflect.DeepEqual(actual, expected) {
		t.Error(expected, actual)
	}
}
func TestRemove(t *testing.T) {
	slice := []int{1, 2, 3, 4}
	slice = Remove(slice, 2).([]int)
	if !reflect.DeepEqual(slice, []int{1, 2, 4}) { t.Error(slice) }
	slice = Remove(slice, 2).([]int)
	if !reflect.DeepEqual(slice, []int{1, 2}) { t.Error(slice) }
	slice = Remove(slice, 0).([]int)
	if !reflect.DeepEqual(slice, []int{2}) { t.Error(slice) }
	slice = Remove(slice, 0).([]int)
	if !reflect.DeepEqual(slice, []int{}) { t.Error(slice) }
}
func TestInArray(t *testing.T) {
	if !InArray(3, []int{1, 2, 3}) { t.Error() }
	if InArray("3", []string{"1", "2"}) { t.Error() }
	//InArray("1", []int{1, 2, 3})
}
func TestDummyMap(t *testing.T) {
	assert := func(condition bool) {
		if !condition {
			Dump(Trace(false)[1])
			t.Error()
		}
	}

	m := DummyMap{}
	assert(!m.Has("test"))
	assert(m.Len() == 0)
	if _, ok := m.Get("test"); ok { t.Fail() }
	assert(reflect.DeepEqual(
		m.Arr(reflect.TypeOf([]string{})),
		[]string{},
	))

	m.Set("test", "a test")
	assert(m.Has("test"))
	assert(m.Len() == 1)
	(func() {
		value, ok := m.Get("test")
		assert(reflect.DeepEqual(value, "a test"))
		assert(ok == true)
	})()
	assert(reflect.DeepEqual(
		m.Arr(reflect.TypeOf([]string{})),
		[]string{"a test"},
	))

	m.Del("test")
	assert(!m.Has("test"))
	assert(m.Len() == 0)
	if _, ok := m.Get("test"); ok { t.Fail() }
	assert(reflect.DeepEqual(
		m.Arr(reflect.TypeOf([]string{})),
		[]string{},
	))
}
func TestTypes(t *testing.T) {
	types := Types(false)
	if !reflect.DeepEqual(
		types,
		[]reflect.Type{
			reflect.TypeOf(Frame{}),
			reflect.TypeOf(Frames{}),
			reflect.TypeOf(ParsedInterface{}),
			reflect.TypeOf(ParsedMethod{}),
			reflect.TypeOf(ParsedPackage{}),
			reflect.TypeOf(ParsedStruct{}),
		},
	) {
		t.Error(types)
	}
}
func TestParseTypes(t *testing.T) {
	parsed := ParseTypes()
	testInterface := parsed.Interfaces["TestInterface"]
	testType1 := parsed.Structs["TestType1"]
	testType2 := parsed.Structs["TestType2"]
	if !reflect.DeepEqual(
		testInterface,
		ParsedInterface{
			Methods: map[string]ParsedMethod{
				"TestMethod": {"func()"},
			},
		},
	) {
		t.Error(testInterface)
	}
	if !reflect.DeepEqual(
		testType1,
		ParsedStruct{
			Methods: map[string]ParsedMethod{},
		},
	) {
		t.Error(testType1)
	}
	if !reflect.DeepEqual(
		testType2,
		ParsedStruct{
			Methods: map[string]ParsedMethod{
				"TestMethod":  {"func()"},
				"testMethod2": {"func() bool"},
			},
		},
	) {
		t.Error(testType2)
	}
}

type TestInterface interface { TestMethod() }
//goland:noinspection GoUnusedExportedType
type TestType1 struct { TestInterface }
//goland:noinspection GoUnusedExportedType
type TestType2 struct { TestInterface }
func (t TestType2) TestMethod() {}
func (t TestType2) testMethod2() bool { return true }
