package my

type DummyMap[K comparable, V any] struct {
	keys   []K
	values []V
}
func (m *DummyMap[K, V]) Set(key K, value V) {
	index := m.index(key)
	if index == -1 {
		m.keys = append(m.keys, key)
		m.values = append(m.values, value)
	} else {
		m.values[index] = value
	}
}
func (m *DummyMap[K, V]) Has(key K) bool {
	return m.index(key) != -1
}
func (m *DummyMap[K, V]) Get(key K) (V, bool) {
	index := m.index(key)
	if index == -1 {
		var res V
		return res, false
	} else {
		return m.values[index], true
	}
}
func (m *DummyMap[K, V]) Del(key K) {
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
func (m *DummyMap[K, V]) Len() int {
	return len(m.keys)
}
func (m *DummyMap[K, V]) Arr() []V {
	if len(m.values) == 0 {
		return []V{}
	} else {
		return m.values
	}
}
func (m *DummyMap[K, V]) index(key K) int {
	for i, k := range m.keys {
		if k == key { return i }
	}
	return -1
}
