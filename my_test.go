package my

import (
	"errors"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestDump(t *testing.T) {
	Dump("hi")
}
func TestTrace(t *testing.T) {
	currentDir, err := os.Getwd()
	PanicIf(err)
	AssertEquals(t, Trace(false), Frames{{currentDir + "/my_test.go", 17}})

	fullTrace := Trace(true)
	Assert(t, len(fullTrace) == 3, fullTrace)
}
func TestSdump2(t *testing.T) {
	AssertEquals(
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
	AssertEquals(t, Revert([]int{1, 2, 3, 4}).([]int), []int{4, 3, 2, 1})
}
func TestRemove(t *testing.T) {
	slice := []int{1, 2, 3, 4}
	slice = Remove(slice, 2).([]int)
	AssertEquals(t, slice, []int{1, 2, 4})
	slice = Remove(slice, 2).([]int)
	AssertEquals(t, slice, []int{1, 2})
	slice = Remove(slice, 0).([]int)
	AssertEquals(t, slice, []int{2})
	slice = Remove(slice, 0).([]int)
	AssertEquals(t, slice, []int{})
}
func TestInArray(t *testing.T) {
	Assert(t, InArray(3, []int{1, 2, 3}))
	Assert(t, !InArray("3", []string{"1", "2"}))
	//InArray("1", []int{1, 2, 3})
}
func TestDummyMap(t *testing.T) {
	m := DummyMap{}
	Assert(t, !m.Has("test"))
	Assert(t, m.Len() == 0)
	if _, ok := m.Get("test"); ok { t.Fail() }
	AssertEquals(
		t,
		m.Arr(reflect.TypeOf([]string{})),
		[]string{},
	)

	m.Set("test", "a test")
	Assert(t, m.Has("test"))
	Assert(t, m.Len() == 1)
	(func() {
		value, ok := m.Get("test")
		AssertEquals(t, value, "a test")
		Assert(t, ok)
	})()
	AssertEquals(
		t,
		m.Arr(reflect.TypeOf([]string{})),
		[]string{"a test"},
	)

	m.Del("test")
	Assert(t, !m.Has("test"))
	Assert(t, m.Len() == 0)
	if _, ok := m.Get("test"); ok { t.Fail() }
	AssertEquals(
		t,
		m.Arr(reflect.TypeOf([]string{})),
		[]string{},
	)
}
func TestTypes(t *testing.T) {
	AssertEquals(
		t,
		Types(false),
		[]reflect.Type{
			reflect.TypeOf(Frame{}),
			reflect.TypeOf(Frames{}),
			reflect.TypeOf(OrderedMapPair{}),
			reflect.TypeOf(ParsedArrayType{}),
			reflect.TypeOf(ParsedChanType{}),
			reflect.TypeOf(ParsedEllipsisType{}),
			reflect.TypeOf(ParsedFuncType{}),
			reflect.TypeOf(ParsedInterface{}),
			reflect.TypeOf(ParsedMapType{}),
			reflect.TypeOf(ParsedNamedType{}),
			reflect.TypeOf(ParsedPackage{}),
			reflect.TypeOf(ParsedStruct{}),
			reflect.TypeOf((*ParsedType)(nil)).Elem(),
		},
	)
}
func TestParseTypes(t *testing.T) {
	parsed := ParseTypes()
	testInterface := parsed.Interfaces()["TestInterface"]
	testType1 := parsed.Structs()["TestType1"]
	testType2 := parsed.Structs()["TestType2"]
	AssertEquals(
		t,
		testInterface,
		ParsedInterface{
			methods: map[string]ParsedFuncType{
				"TestMethod": {},
			},
		},
	)
	AssertEquals(
		t,
		testType1,
		ParsedStruct{
			methods: map[string]ParsedFuncType{},
		},
	)
	AssertEquals(
		t,
		testType2,
		ParsedStruct{
			methods: map[string]ParsedFuncType{
				"TestMethod":  {},
				"testMethod2": {out: []ParsedType{ParsedNamedType{"bool"}}},
			},
		},
	)

	Assert(t, !testType1.Implements(testInterface))
	Assert(t, testType2.Implements(testInterface))
}
func TestStartCommand(t *testing.T) {
	outSlice := make([]string, 0)
	cmd := StartCommand(
		"",
		"sleep 0.1 && echo 1 && sleep 0.1 && echo 2",
		func(out string) { outSlice = append(outSlice, out) },
		func(err string) { panic(err) },
	)
	result := make(chan error)
	go func() {
		result <- cmd.Wait()
		close(result)
	}()
	for _, testCase := range []struct {
		sleep       time.Duration
		exited      bool
		expectedOut []string
	}{
		{90 * time.Millisecond, false, []string{}},
		{20 * time.Millisecond, false, []string{"1"}},
		{80 * time.Millisecond, false, []string{"1"}},
		{20 * time.Millisecond, true, []string{"1", "2"}},
	} {
		time.Sleep(testCase.sleep)
		select {
			case status := <- result: Assert(t, status == nil)
			default: Assert(t, !testCase.exited)
		}
		AssertEquals(t, outSlice, testCase.expectedOut)
	}
}
func TestOrderedMap(t *testing.T) {
	m := OrderedMap{}.New()
	checkMap := func(lenKeys, lenValues, lenIndices, lenDeletedIndices int) {
		Assert(t, len(m.keys) == lenKeys)
		Assert(t, len(m.values) == lenValues)
		Assert(t, len(m.indices) == lenIndices)
		Assert(t, len(m.deletedIndices) == lenDeletedIndices)
	}
	checkMap(0, 0, 0, 0)
	m.Add("key1", "value1")
	m.Add("key2", "value2")
	m.Add("key3", "value3")
	m.Add("key4", "value4")
	m.Add("key5", "value5")
	m.Add("key6", "value6")
	m.Add("key7", "value7")
	checkMap(7, 7, 7, 0)
	AssertEquals(
		t,
		m.Pairs(),
		[]OrderedMapPair{
			{"key1", "value1"},
			{"key2", "value2"},
			{"key3", "value3"},
			{"key4", "value4"},
			{"key5", "value5"},
			{"key6", "value6"},
			{"key7", "value7"},
		},
	)
	m.Del("key2")
	checkMap(7, 7, 6, 1)
	m.Del("key6")
	m.Del("key4")
	checkMap(7, 7, 4, 3)
	AssertEquals(
		t,
		m.Pairs(),
		[]OrderedMapPair{
			{"key1", "value1"},
			{"key3", "value3"},
			{"key5", "value5"},
			{"key7", "value7"},
		},
	)
	m.Del("key7")
	checkMap(3, 3, 3, 0)
	AssertEquals(
		t,
		m.Pairs(),
		[]OrderedMapPair{
			{"key1", "value1"},
			{"key3", "value3"},
			{"key5", "value5"},
		},
	)
}

type TestInterface interface { TestMethod() }
//goland:noinspection GoUnusedExportedType
type TestType1 struct { TestInterface }
//goland:noinspection GoUnusedExportedType
type TestType2 struct { TestInterface }
func (t TestType2) TestMethod() {}
func (t TestType2) testMethod2() bool { return true }
