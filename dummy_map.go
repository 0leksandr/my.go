package my

import "reflect"

type DummyMap struct {
	keys   []string
	values []interface{}
}
func (m *DummyMap) Set(key string, value interface{}) {
	index := m.index(key)
	if index == -1 {
		m.keys = append(m.keys, key)
		m.values = append(m.values, value)
	} else {
		m.values[index] = value
	}
}
func (m *DummyMap) Has(key string) bool {
	return m.index(key) != -1
}
func (m *DummyMap) Get(key string) (interface{}, bool) {
	index := m.index(key)
	if index == -1 {
		return nil, false
	} else {
		return m.values[index], true
	}
}
func (m *DummyMap) Del(key string) {
	index := m.index(key)
	if index != -1 {
		last := len(m.keys) - 1
		if index != last {
			m.keys[index] = m.keys[last]
			m.values[index] = m.values[last]
		}
		m.keys = m.keys[:last]
		m.values = m.values[:last]
	}
}
func (m *DummyMap) Len() int {
	return len(m.keys)
}
func (m *DummyMap) Arr(arrType reflect.Type) interface{} {
	arr := reflect.MakeSlice(arrType, 0, m.Len())
	for _, value := range m.values {
		arr = reflect.Append(arr, reflect.ValueOf(value))
	}
	return arr.Interface()
}
func (m *DummyMap) index(key string) int {
	for i, k := range m.keys {
		if k == key { return i }
	}
	return -1
}
