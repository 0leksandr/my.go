package my

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"sync"
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
func TestInsertAt(t *testing.T) {
	slice := []string{"1", "2", "3"}
	AssertEquals(t, InsertAt(slice, 1, "4"), []string{"1", "4", "2", "3"})
	AssertEquals(t, slice, []string{"1", "2", "3"})
	slice = InsertAt(slice, 0, "4")
	AssertEquals(t, slice, []string{"4", "1", "2", "3"})
	slice = InsertAt(slice, 4, "5")
	AssertEquals(t, slice, []string{"4", "1", "2", "3", "5"})
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
	parsedPackages := ParseTypes()
	AssertEquals(t, len(parsedPackages), 1)
	parsed := parsedPackages["my"]

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
			embedded: []ParsedNamedType{{"TestInterface", nil}},
			methods:  map[string]ParsedFuncType{},
		},
	)
	Assert(t, !testType1.Implements(testInterface))

	testType2 := parsed.Structs()["TestType2"]
	AssertEquals(
		t,
		testType2,
		ParsedStruct{
			embedded: []ParsedNamedType{{"TestInterface", nil}},
			methods: map[string]ParsedFuncType{
				"TestMethod":  {},
				"testMethod2": {out: []ParsedType{ParsedNamedType{"bool", nil}}},
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
		reflect.TypeOf((*Clock)(nil)).Elem(),
		reflect.TypeOf(Error{}),
		reflect.TypeOf(Frame{}),
		reflect.TypeOf(Locker{}),
		reflect.TypeOf(MockClock{}),
		reflect.TypeOf(ParsedArrayType{}),
		reflect.TypeOf(ParsedChanType{}),
		reflect.TypeOf(ParsedEllipsisType{}),
		reflect.TypeOf(ParsedFuncType{}),
		reflect.TypeOf(ParsedIndex{}),
		reflect.TypeOf(ParsedInterface{}),
		reflect.TypeOf(ParsedMapType{}),
		reflect.TypeOf(ParsedNamedType{}),
		reflect.TypeOf(ParsedStruct{}),
		reflect.TypeOf((*ParsedType)(nil)).Elem(),
		reflect.TypeOf(RealClock{}),
		reflect.TypeOf(TestTrend{}),
		reflect.TypeOf((*TimerInterface)(nil)).Elem(),
		reflect.TypeOf(Trace{}),
		reflect.TypeOf((*Trend)(nil)).Elem(),
	}
	for _, expectedType := range expectedTypes { Assert(t, InArray(expectedType, actualTypes)) }

	AssertEquals(
		t,
		ArrayMap(actualTypes, func(t reflect.Type) string { return t.String() }),
		[]string{
			"my.Call",
			"my.Clock",
			//"my.DummyMap[string,string]",
			"my.Error",
			"my.Expected",
			"my.Frame",
			"my.K",
			"my.Locker",
			"my.MockClock",
			"my.OrderedMapPair[github.com/0leksandr/my%2ego.K·3,github.com/0leksandr/my%2ego.V·4]",
			"my.OrderedMap[github.com/0leksandr/my%2ego.K·3,github.com/0leksandr/my%2ego.V·4]",
			"my.OrderedMap[go.shape.string,go.shape.string]",
			"my.ParsedArrayType",
			"my.ParsedChanType",
			"my.ParsedEllipsisType",
			"my.ParsedFuncType",
			"my.ParsedIndex",
			"my.ParsedInterface",
			"my.ParsedMapType",
			"my.ParsedNamedType",
			"my.ParsedStruct",
			"my.ParsedType",
			"my.RealClock",
			"my.ReservoirQueue[github.com/0leksandr/my%2ego.zeroQueueReceiver[int]]",
			"my.ReservoirQueue[int]",
			"my.TestCase",
			"my.TestTrend",
			"my.TimerInterface",
			"my.Trace",
			"my.Trend",
			"my.V",
			"my.ZeroQueue[int]",
			"my.event",
			"my.funcMockTimerCallee",
			"my.initMapsTest",
			"my.key",
			"my.mockTimer",
			"my.mockTimerCallee",
			"my.testReference",
			"my.testString",
			"my.zeroQueueReceiver[go.shape.int]",
			"my.zeroQueueReceiver[int]",
			"my.zeroQueueZeroReceiver[go.shape.int]",
			"my.zeroQueueZeroReceiver[int]",
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
			case status := <- result: Assert(t, status == nil, testCase)
			default: Assert(t, !testCase.exited, testCase)
		}
		AssertEquals(t, outSlice, testCase.expectedOut, testCase)
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
			//{error1A, error2A},
			{error1A, error3B},
			{error2A, error3B},
		} {
			Assert(t, !errors.Is(pair[0], pair[1]))
			Assert(t, !errors.Is(pair[1], pair[0]))
		}
	}
}
func TestComboError(t *testing.T) {
	AssertEquals(t, ComboError(nil, nil), nil)
	AssertEquals(t, ComboError(nil, nil, nil, nil, nil), nil)
	AssertEquals(t, ComboError(errors.New("test"), nil), errors.New("test"))
	AssertEquals(t, ComboError(nil, errors.New("test"), nil), errors.New("test"))
	err1 := errors.New("1")
	err2 := errors.New("2")
	err3 := errors.New("3")
	AssertEquals(t, ComboError(err1, nil, err2), Error{}.New("1\n---\n2"))
	AssertEquals(t, ComboError(err1, nil, err2, nil, err3, nil), Error{}.New("1\n---\n2\n---\n3"))
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
func TestLocker(t *testing.T) {
	var locker Locker

	var events []string
	var accessEvents sync.Mutex
	write := func(event string) {
		accessEvents.Lock()
		events = append(events, event)
		accessEvents.Unlock()
	}
	lock := func() {
		write("locking")
		locker.Lock()
		write("locked")
	}
	unlock := func() {
		write("unlocking")
		locker.Unlock()
		write("unlocked")
	}
	wait := func() {
		write("waiting")
		locker.Wait()
		write("waited")
	}

	wait()
	lock()
	unlock()
	unlock()
	lock()
	lock()

	nrWaits := 2
	var startWaiting, doneWaiting sync.WaitGroup
	startWaiting.Add(nrWaits)
	doneWaiting.Add(nrWaits)
	for i := 0; i < nrWaits; i++ {
		go func() {
			startWaiting.Done()
			wait()
			doneWaiting.Done()
		}()
	}
	go func() {
		startWaiting.Wait()
		time.Sleep(1 * time.Microsecond)
		lock()
		unlock()
	}()

	doneWaiting.Wait()

	AssertEquals(t, events, []string{
		"waiting", "waited",
		"locking", "locked",
		"unlocking", "unlocked",
		"unlocking", "unlocked",
		"locking", "locked",
		"locking", "locked",
		"waiting", "waiting",
			"locking", "locked",
			"unlocking", "unlocked",
		"waited", "waited",
	})
}
func TestReservoir(t *testing.T) {
	in := make(chan int)
	out := Reservoir(in, 10)
	values := []int{5, 7766, 232, 65, -23}
	for _, value := range values { in <- value }
	close(in)
	var result []int
	for value := range out { result = append(result, value) }
	AssertEquals(t, result, values)
}
func TestDispenser(t *testing.T) {
	var processed [][]int
	channel := make(chan int, 10)
	var locker1, locker2 Locker

	locker1.Lock()
	go Dispenser(channel, func(values []int) {
		locker1.Wait()
		processed = append(processed, values)
		locker2.Unlock()
	})

	for i := 0; i < 3; i++ { channel <- i }
	locker2.Lock()
	locker1.Unlock()
	locker2.Wait()

	locker1.Lock()
	for i := 3; i < 6; i++ { channel <- i }
	locker2.Lock()
	locker1.Unlock()
	locker2.Wait()

	AssertEquals(t, processed, [][]int{{0, 1, 2}, {3, 4, 5}})
}
func TestReservoirQueue(t *testing.T) {
	//queue := (*FairChannelQueue[int])(nil).New()
	//queue := ChannelQueue[int]{}.New(3)
	queue := ReservoirQueue[int]{}.New()
	var values []int
	var accessValues sync.Mutex
	var wg sync.WaitGroup
	get := func() {
		wg.Add(1)
		go func() {
			accessValues.Lock()
			values = append(values, queue.Get())
			accessValues.Unlock()
			wg.Done()
		}()
	}
	for _, i := range []int{1, 2, 3} { queue.Put(i) }
	for i := 0; i < 7; i++ {
		time.Sleep(1 * time.Millisecond)
		get()
	}
	time.Sleep(1 * time.Millisecond)
	for _, i := range []int{4, 5, 6, 7} { queue.Put(i) }
	wg.Wait()
	AssertEquals(t, values, []int{1, 2, 3, 4, 5, 6, 7})
}
func TestZeroQueue(t *testing.T) {
	queue := ZeroQueue[int]{}.New()
	var values []string
	var accessReceived sync.Mutex
	store := func(operation string, value int) {
		accessReceived.Lock()
		values = append(values, fmt.Sprintf("%s %d", operation, value))
		accessReceived.Unlock()
	}
	var wg sync.WaitGroup
	const NrItems = 5
	for i := 0; i < NrItems; i++ {
		wg.Add(2)
		go func(i int) {
			time.Sleep(time.Duration(i) * 10 * time.Millisecond)
			queue.Put(i)
			store("put", i)
			wg.Done()
		}(i)
		go func() {
			value := queue.Get()
			store("got", value)
			wg.Done()
		}()
	}
	wg.Wait()
	time.Sleep(10 * time.Millisecond) // MAYBE: find a better way to deal with race condition
	AssertEquals(t, values, []string{
		"got 0",
		"put 0",
		"got 1",
		"put 1",
		"got 2",
		"put 2",
		"got 3",
		"put 3",
		"got 4",
		"put 4",
	})
}
func TestMockTime(t *testing.T) {
	MockTime(func(mockClock *MockClock) {
		start := clock.Now()
		type event struct {
			title string
			time  time.Time
		}
		var events []event
		assertEvents := func(expected []event) {
			time.Sleep(10 * time.Millisecond) // MAYBE: find a better way to deal with race condition
			AssertEquals(t, events, expected)
			events = nil
		}

		nrSleeps := 5
		var wg sync.WaitGroup
		wg.Add(nrSleeps)
		for i := 1; i <= nrSleeps; i++ {
			go func(i int) {
				mockClock.AfterFunc(time.Duration(i) * time.Second, func() {
					events = append(events, event{fmt.Sprintf("slept for %ds", i), mockClock.Now()})
				})
				wg.Done()
			}(i)
		}
		wg.Wait()

		clock.Sleep(999 * time.Millisecond)
		assertEvents(nil)

		clock.Sleep(1 * time.Millisecond)
		assertEvents([]event{{"slept for 1s", start.Add(1 * time.Second)}})

		clock.Sleep(999 * time.Millisecond)
		assertEvents(nil)

		clock.Sleep(1 * time.Millisecond)
		assertEvents([]event{{"slept for 2s", start.Add(2 * time.Second)}})

		clock.Sleep(10 * time.Second)
		assertEvents([]event{
			{"slept for 3s", start.Add(3 * time.Second)},
			{"slept for 4s", start.Add(4 * time.Second)},
			{"slept for 5s", start.Add(5 * time.Second)},
		})
	})
}
func TestDelayer(t *testing.T) { // TODO: move
	ms := func(ms int) time.Duration {
		return time.Duration(ms) * time.Millisecond
	}
	sequence := func(from, to int) []int {
		var sequence []int
		for i := from; i <= to; i++ { sequence = append(sequence, i) }
		return sequence
	}
	multiplyEach := func(sequence []int, n int) []int {
		multiplied := make([]int, len(sequence))
		for i, v := range sequence { multiplied[i] = v * n }
		return multiplied
	}
	type Expected struct {
		time    int
		nrTicks int
	}
	type TestCase struct {
		min      int
		max      int
		events   []int
		expected []Expected
	}
	type Call struct {
		duration time.Duration
		nrTicks  int
	}

	testCases := []TestCase{
		{
			10,
			20,
			[]int{
				7,
				8,
				12, //  5 + 7  | 27
				18, // 10 + 8  | 45
				22, // 10 + 12 | 67
				16, // 10 + 6  | 83
				8,
			},
			[]Expected{
				{20, 3},
				{37, 1},
				{55, 1},
				{77, 1},
				{91, 2}, // 101
			},
		},
		{
			5,
			15,
			multiplyEach(sequence(0, 6), 2),
			[]Expected{
				{11, 4}, // 0 + 2 + 4 + (5 | 1)
				{17, 1}, // (5 | 3)
				{25, 1}, // (5 | 5)
				{35, 1}, // (5 | 7)
				{42, 1},
			},
		},
		{
			10,
			15,
			sequence(0, 10),
			[]Expected{
				{15, 6},
				{30, 3},
				{51, 2},
				{55, 1},
			},
		},
		{
			9,
			15,
			multiplyEach(sequence(0, 6), 2),
			[]Expected{
				{15, 5}, // (3 | 5)
				{29, 1}, // (9 | 1)
				{39, 1}, // (9 | 3)
				{42, 1},
			},
		},
	}
	for _, testCase := range testCases {
		ticks := make(chan struct{})
		var calls []Call
		var accessCalls sync.Mutex
		MockTime(func(clock *MockClock) {
			start := clock.Now()
			go func() {
				ticks <- struct{}{}
				for _, duration := range testCase.events {
					time.Sleep(10 * time.Millisecond)
					clock.Sleep(ms(duration))
					time.Sleep(10 * time.Millisecond)
					ticks <- struct{}{}
				}
				close(ticks)
			}()
			Delayer(ticks, ms(testCase.min), ms(testCase.max), func(ticks []struct{}) {
				accessCalls.Lock()
				calls = append(calls, Call{clock.Now().Sub(start), len(ticks)})
				accessCalls.Unlock()
			})
		})
		AssertEquals(
			t,
			calls,
			ArrayMap(testCase.expected, func(expected Expected) Call {
				return Call{ms(expected.time), expected.nrTicks}
			}),
		)
	}
}
func TestInitMaps(t *testing.T) {
	type initMapsTest struct {
		map1 map[string]string
		map2 map[string]map[int]float32
		//ptr1 *map[string]int
		//str1 *initMapsTest
	}
	type testCase struct {
		test     initMapsTest
		expected initMapsTest
	}
	for _, _testCase := range []testCase{
		{
			initMapsTest{},
			initMapsTest{
				map[string]string{},
				map[string]map[int]float32{},
				//&map[string]int{},
				//nil,
			},
		},
	}{
		AssertEquals(t, InitMaps(_testCase.test), _testCase.expected)
	}
}
func TestFormatter(t *testing.T) {
	type testString struct {
		value string
	}

	type testReference struct {
		reference *testReference
	}

	type testCase struct {
		value             any
		expectedFormatted string
	}
	for _, _testCase := range []testCase{
		{ "hello",           `"hello"`           },
		{ nil,               "nil"               },
		{ true,              "true"              },
		{ false,             "false"             },
		{ -13,               "-13"               },
		{ 6.66,              "6.66"              },
		{ -1.23e+45,         "-1.23e+45"         },
		//{ 1e100000,          "INF"               },
		{ "123",             `"123"`             },
		{ 123,               "123"               },
		{ 123.,              "123."              },
		{ ([]int)(nil),      "([]int)(nil)"      },
		{ []int{},           "[]int{}"           },
		{ (chan int)(nil),   "(chan int)(nil)"   },
		{ (<-chan int)(nil), "(<-chan int)(nil)" },
		{ (*int)(nil),       "(*int)(nil)"       },
		{ []byte("test"),    `[]byte("test")`    },
		{
			func(int) (_ string) { return },
			"func(int) (_ string) { return }",
		},
		{
			(func(int) string)(nil),
			"(func(int) string)(nil)",
		},
		{
			func(int, float64) (string, error) { return "", nil },
			"func(int, float64) (_ string, _ error) { return }",
		},
		{
			[]int{1, 2, 3},
`[]int{
	1,
	2,
	3,
}`,
		},
		{
			testString{
				value: "hello",
			},
`my.testString{
	value: "hello",
}`,
		},
		{
			[]any{
				"hello",
				"\"hello\"",
				0,
				0.,
			},
`[]interface {}{
	"hello",
	"\"hello\"",
	0,
	0.,
}`,
		},
		{
`
	this is a
	multiline text
`,
"`" + `
	this is a
	multiline text
` + "`",
		},
		{
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
"`" + `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempo
r incididunt ut labore et dolore magna aliqua.` + "`",
		},
		{(map[string]int)(nil), "(map[string]int)(nil)"},
		{map[string]int{}, "map[string]int{}"},
		{
			map[string]any{
				"test1": 2,
				"test3": "4",
			},
`map[string]interface {}{
	"test1": 2,
	"test3": "4",
}`,
		},
		{
			map[any]any{
				"test1": 2,
				3:       "4",
				33:      8,
				"33":    "8",
				34:      9,
				5.6:     testString{"7"},
				"func":  func(int, float64) (string, error) { return "", nil },
			},
`map[interface {}]interface {}{
	3: "4",
	5.6: my.testString{
		value: "7",
	},
	33: 8,
	34: 9,
	"33": "8",
	"func": func(int, float64) (_ string, _ error) { return },
	"test1": 2,
}`,
		},
	} {
		AssertEquals(
			t,
			formatter{}.New().format(_testCase.value, 1),
			_testCase.expectedFormatted,
		)
	}

	for _, _testCase := range []testCase{
		{
			(func() *testReference {
				ref0 := &testReference{}
				ref1 := &testReference{ref0}
				ref0.reference = ref1
				return ref0
			})(),
`^&my\.testReference{
	reference: &my\.testReference{
		reference: \[circular reference 0x[0-9a-f]{11}\],
	},
}$`,
		},
	} {
		actual := formatter{}.New().format(_testCase.value, 1)
		Assert(
			t,
			regexp.
				MustCompile(_testCase.expectedFormatted).
				MatchString(actual),
			actual,
		)
	}
}
func TestGetExportedFields(t *testing.T) {
	type TestType struct {
		Field1 string
		field2 string
		Field3 string
	}
	testValue := TestType{
		Field1: "value1",
		field2: "value2",
		Field3: "value3",
	}
	AssertEquals(
		t,
		GetExportedFields(testValue),
		[]any{
			"value1",
			"value3",
		},
	)
}

func TestYo(t *testing.T) {
	t.SkipNow()
	//parseGo(func(decl ast.Decl) {})
	//parseGoDir("/Users/oleksandr.boiko/_/MythicalGames/platform-items", func(decl ast.Decl) {})
	//Dump2(parseTypesRecursively("/Users/oleksandr.boiko/_/MythicalGames/platform-items"))
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
