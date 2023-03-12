package my

import (
	"errors"
	"reflect"
	"runtime"
	"sort"
	"testing"
	"time"
)

func TestDump(t *testing.T) {
	Dump("hi")
}
func TestGetTrace(t *testing.T) {
	_, _, thisLine, _ := runtime.Caller(0)
	AssertEquals(t, GetTrace(false), Trace{{"my_test.go", thisLine + 1}})

	fullTrace := GetTrace(true)
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
	AssertEquals(t, Revert([]int{1, 2, 3, 4}), []int{4, 3, 2, 1})
}
func TestRemove(t *testing.T) {
	slice := []int{1, 2, 3, 4}
	slice = Remove(slice, 2)
	AssertEquals(t, slice, []int{1, 2, 4})
	slice = Remove(slice, 2)
	AssertEquals(t, slice, []int{1, 2})
	slice = Remove(slice, 0)
	AssertEquals(t, slice, []int{2})
	slice = Remove(slice, 0)
	AssertEquals(t, slice, []int{})
}
func TestInArray(t *testing.T) {
	Assert(t, InArray(3, []int{1, 2, 3}))
	Assert(t, !InArray("3", []string{"1", "2"}))
}
func TestKeys(t *testing.T) {
	actualKeys := Keys(map[string]int{"one": 1, "two": 2, "three": 3})
	expectedKeys := []string{"one", "two", "three"}
	sort.Strings(actualKeys)
	sort.Strings(expectedKeys)
	AssertEquals(t, actualKeys, expectedKeys)
}
func TestArrayFilter(t *testing.T) {
	AssertEquals(
		t,
		ArrayFilter([]int{1, 2, 3, 4, 5, 6}, func(i int) bool {
			return i % 2 == 1
		}),
		[]int{1, 3, 5},
	)
}
func TestArrayMap(t *testing.T) {
	AssertEquals(
		t,
		ArrayMap([]string{"one", "two", "three"}, func(s string) string { return "[" + s + "]" }),
		[]string{"[one]", "[two]", "[three]"},
	)
}
func TestDummyMap(t *testing.T) {
	m := DummyMap[string, string]{}
	Assert(t, !m.Has("test"))
	Assert(t, m.Len() == 0)
	if _, ok := m.Get("test"); ok { t.Fail() }
	AssertEquals(
		t,
		m.Arr(),
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
		m.Arr(),
		[]string{"a test"},
	)

	m.Del("test")
	Assert(t, !m.Has("test"))
	Assert(t, m.Len() == 0)
	if _, ok := m.Get("test"); ok { t.Fail() }
	AssertEquals(
		t,
		m.Arr(),
		[]string{},
	)
}
func TestParseTypes(t *testing.T) {
	parsed := ParseTypes()

	testInterface := parsed.Interfaces()["TestInterface"]
	AssertEquals(
		t,
		testInterface,
		ParsedInterface{
			methods: map[string]ParsedFuncType{
				"TestMethod": {},
			},
		},
	)

	testType1 := parsed.Structs()["TestType1"]
	AssertEquals(
		t,
		testType1,
		ParsedStruct{
			embedded: []string{"TestInterface"},
			methods:  map[string]ParsedFuncType{},
		},
	)
	Assert(t, !testType1.Implements(testInterface))

	testType2 := parsed.Structs()["TestType2"]
	AssertEquals(
		t,
		testType2,
		ParsedStruct{
			embedded: []string{"TestInterface"},
			methods: map[string]ParsedFuncType{
				"TestMethod":  {},
				"testMethod2": {out: []ParsedType{ParsedNamedType{"bool"}}},
			},
		},
	)
	Assert(t, testType2.Implements(testInterface))
}
func TestRuntimeTypes(t *testing.T) {
	actualTypes := Types(false)

	// testing all types this way is inconvenient, because:
	// - a type might be expected only because it is added here. This is a test, that influences actual code
	// - some types are hard to add/catch/simulate (`my.OrderedMap[go.shape.string,go.shape.string]`)
	expectedTypes := []reflect.Type{
		reflect.TypeOf(Error{}),
		reflect.TypeOf(Frame{}),
		reflect.TypeOf(ParsedArrayType{}),
		reflect.TypeOf(ParsedChanType{}),
		reflect.TypeOf(ParsedEllipsisType{}),
		reflect.TypeOf(ParsedFuncType{}),
		reflect.TypeOf(ParsedInterface{}),
		reflect.TypeOf(ParsedMapType{}),
		reflect.TypeOf(ParsedNamedType{}),
		reflect.TypeOf(ParsedStruct{}),
		reflect.TypeOf((*ParsedType)(nil)).Elem(),
		reflect.TypeOf(TestTrend{}),
		reflect.TypeOf(Trace{}),
		reflect.TypeOf((*Trend)(nil)).Elem(),
	}
	for _, expectedType := range expectedTypes { Assert(t, InArray(expectedType, actualTypes)) }

	AssertEquals(
		t,
		ArrayMap(actualTypes, func(t reflect.Type) string { return t.String() }),
		[]string{
			"my.DummyMap[string,string]",
			"my.Error",
			"my.Frame",
			"my.K",
			"my.OrderedMapPair[github.com/0leksandr/my%2ego.K·2,github.com/0leksandr/my%2ego.V·3]",
			"my.OrderedMap[github.com/0leksandr/my%2ego.K·2,github.com/0leksandr/my%2ego.V·3]",
			"my.OrderedMap[go.shape.string,go.shape.string]",
			"my.ParsedArrayType",
			"my.ParsedChanType",
			"my.ParsedEllipsisType",
			"my.ParsedFuncType",
			"my.ParsedInterface",
			"my.ParsedMapType",
			"my.ParsedNamedType",
			"my.ParsedStruct",
			"my.ParsedType",
			"my.TestTrend",
			"my.Trace",
			"my.Trend",
			"my.V",
		},
	)
}
func TestLocalTypes(t *testing.T) {
	testTypes(t, []string{"TestType1"})
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
	type K string
	type V string
	type Pair struct{k K; v V}
	m := (*OrderedMap[K, V])(nil).New()
	assertMap := func(
		m *OrderedMap[K, V],
		lenKeys,
		lenValues,
		lenIndices,
		lenDeletedIndices int,
		values []Pair,
	) {
		Assert(t, len(m.keys) == lenKeys)
		Assert(t, len(m.values) == lenValues)
		Assert(t, len(m.indices) == lenIndices)
		Assert(t, len(m.deletedIndices) == lenDeletedIndices)
		assertValues := func(m *OrderedMap[K, V]) {
			var index int
			m.Each(func(key K, value V) {
				row := values[index]
				index++
				AssertEquals(t, key, row.k)
				AssertEquals(t, value, row.v)
			})
		}
		assertValues(m)

		c := m.Copy()
		l := len(values)
		Assert(t, len(c.keys) == l)
		Assert(t, len(c.values) == l)
		Assert(t, len(c.indices) == l)
		Assert(t, len(c.deletedIndices) == 0)
		assertValues(c)
	}
	assertMap(m, 0, 0, 0, 0, []Pair{})
	m.Add("key1", "value1")
	m.Add("key2", "value2")
	m.Add("key3", "value3")
	m.Add("key4", "value4")
	m.Add("key5", "value5")
	m.Add("key6", "value6")
	m.Add("key7", "value7")
	assertMap(m, 7, 7, 7, 0, []Pair{
		{"key1", "value1"},
		{"key2", "value2"},
		{"key3", "value3"},
		{"key4", "value4"},
		{"key5", "value5"},
		{"key6", "value6"},
		{"key7", "value7"},
	})
	AssertEquals(
		t,
		m.Pairs(),
		[]OrderedMapPair[K, V]{
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
	assertMap(m, 7, 7, 6, 1, []Pair{
		{"key1", "value1"},
		{"key3", "value3"},
		{"key4", "value4"},
		{"key5", "value5"},
		{"key6", "value6"},
		{"key7", "value7"},
	})
	m.Del("key6")
	m.Del("key4")
	assertMap(m, 7, 7, 4, 3, []Pair{
		{"key1", "value1"},
		{"key3", "value3"},
		{"key5", "value5"},
		{"key7", "value7"},
	})
	AssertEquals(
		t,
		m.Pairs(),
		[]OrderedMapPair[K, V]{
			{"key1", "value1"},
			{"key3", "value3"},
			{"key5", "value5"},
			{"key7", "value7"},
		},
	)

	m2 := (*OrderedMap[K, V])(nil).New()
	Assert(
		t,
		m.Equals(
			m2.
				Add("key1", "value1").
				Add("key3", "value3").
				Add("key5", "value5").
				Add("key7", "value7"),
		),
	)
	m2 = (*OrderedMap[K, V])(nil).New()
	Assert(
		t,
		!m.Equals(
			m2.
				Add("key1", "value1").
				Add("key3", "value3").
				Add("key5", "value5").
				Add("key7", "value8"),
		),
	)

	m.Del("key7")
	assertMap(m, 3, 3, 3, 0, []Pair{
		{"key1", "value1"},
		{"key3", "value3"},
		{"key5", "value5"},
	})
	AssertEquals(
		t,
		m.Pairs(),
		[]OrderedMapPair[K, V]{
			{"key1", "value1"},
			{"key3", "value3"},
			{"key5", "value5"},
		},
	)

	m3 := (*OrderedMap[K, V])(nil).New()
	m3.Add("key1", "value1")
	m3.Add("key2", "value2")
	m3.Del("key1")
	m3.Add("key1", "value3")
	m3.Del("key2")
	key1, _ := m3.Get("key1")
	AssertEquals(t, key1, V("value3"))
}
func TestError(t *testing.T) {
	_, _, thisLine, _ := runtime.Caller(0)
	f := func() error {
		return Error{}.New("test")
	}
	AssertEquals(
		t,
		f(),
		Error{
			error: errors.New("test"),
			trace: Trace{
				{"my_test.go", thisLine + 2},
				{"my_test.go", thisLine + 6},
			},
		},
	)
}
func TestError_WrapUnwrap(t *testing.T) {
	for _, underlying := range [][2]error{
		{
			errors.New("testA"),
			errors.New("testB"),
		},
		{
			Error{}.New("testA"),
			Error{}.New("testB"),
		},
	} {
		testA := underlying[0]
		testB := underlying[1]
		error1A := Error{}.Wrap(testA)
		error2A := Error{}.Wrap(testA)
		error3B := Error{}.Wrap(testB)

		Assert(t, errors.Is(error1A, testA))
		Assert(t, errors.Is(error2A, testA))
		Assert(t, errors.Is(error3B, testB))
		Assert(t, !errors.Is(error1A, testB))
		Assert(t, !errors.Is(error2A, testB))
		Assert(t, !errors.Is(error3B, testA))

		for _, pair := range [][2]Error{
			{error1A, error2A},
			{error1A, error3B},
			{error2A, error3B},
		} {
			Assert(t, !errors.Is(pair[0], pair[1]))
			Assert(t, !errors.Is(pair[1], pair[0]))
		}
	}
}
func TestProgressBar(t *testing.T) {
	if false {
		testProgress()
	}
}
func TestTopChart(t *testing.T) {
	for _, testCase := range[][]struct{i int; expected []int}{
		{
			{1, []int{1}},
			{2, []int{2, 1}},
			{3, []int{3, 2, 1}},
			{4, []int{4, 3, 2}},
			{5, []int{5, 4, 3}},
		},
		{
			{5, []int{5}},
			{4, []int{5, 4}},
			{3, []int{5, 4, 3}},
			{2, []int{5, 4, 3}},
			{1, []int{5, 4, 3}},
		},
		{
			{3, []int{3}},
			{1, []int{3, 1}},
			{4, []int{4, 3, 1}},
			{2, []int{4, 3, 2}},
			{5, []int{5, 4, 3}},
		},
	} {
		topChart := (*TopChart)(nil).OfConstSize(3)
		for _, testCase2 := range testCase {
			topChart.Add(TestTrend{testCase2.i})
			actual := make([]int, 0, 3)
			for _, trend := range topChart.Trends {
				actual = append(actual, trend.(TestTrend).value)
			}
			AssertEquals(t, actual, testCase2.expected)
		}
	}
}

type TestInterface interface { TestMethod() }
//goland:noinspection GoUnusedExportedType
type TestType1 struct { TestInterface }
//goland:noinspection GoUnusedExportedType
type TestType2 struct { TestInterface }
func (t TestType2) TestMethod() {}
func (t TestType2) testMethod2() bool { return true }

type TestTrend struct {
	value int
}
func (trend TestTrend) TrendValue() float64 {
	return float64(trend.value)
}
