package my

type OrderedMap[K comparable, V comparable] struct {
	keys           []K
	values         []V
	indices        map[K]int
	deletedIndices map[int]bool
}
type OrderedMapPair[K comparable, V any] struct {
	Key   K
	Value V
}
func (*OrderedMap[K, V]) New() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		keys:           make([]K, 0),
		values:         make([]V, 0),
		indices:        make(map[K]int),
		deletedIndices: make(map[int]bool),
	}
}
func (m *OrderedMap[K, V]) Add(key K, value V) *OrderedMap[K, V] {
	if m.Has(key) { panic("key already filled") }
	m.keys = append(m.keys, key)
	m.values = append(m.values, value)
	m.indices[key] = len(m.keys) - 1
	return m
}
func (m *OrderedMap[K, V]) Len() int {
	return len(m.indices)
}
func (m *OrderedMap[K, V]) Get(key K) (V, bool) {
	if index, ok := m.indices[key]; ok {
		return m.values[index], true
	} else {
		var res V
		return res, false
	}
}
func (m *OrderedMap[K, V]) Has(key K) bool {
	_, ok := m.indices[key]
	return ok
}
func (m *OrderedMap[K, V]) Del(key K) {
	if index, ok := m.indices[key]; ok {
		m.deletedIndices[index] = true
		delete(m.indices, m.keys[index])

		if len(m.deletedIndices) > len(m.indices) {
			l := len(m.indices) + len(m.deletedIndices)
			indices := make([]int, l)
			for i := 0; i < l; i++ { indices[i] = i }
			for _index := range m.deletedIndices { indices[_index] = -1 }

			keys := make([]K, 0, len(m.indices))
			values := make([]V, 0, len(m.indices))
			j := 0
			for _, _index := range indices {
				if _index != -1 {
					key2 := m.keys[_index]
					keys = append(keys, key2)
					values = append(values, m.values[_index])
					m.indices[key2] = j
					j++
				}
			}

			m.keys = keys
			m.values = values
			m.deletedIndices = make(map[int]bool)
		}
	}
}
func (m *OrderedMap[K, V]) Each(f func(key K, value V)) {
	for index, key := range m.keys {
		if !m.deletedIndices[index] {
			f(key, m.values[index])
		}
	}
}
func (m *OrderedMap[K, V]) Pairs() []OrderedMapPair[K, V] {
	pairs := make([]OrderedMapPair[K, V], 0, m.Len())
	m.Each(func(key K, value V) {
		pairs = append(
			pairs,
			OrderedMapPair[K, V]{
				Key:   key,
				Value: value,
			},
		)
	})
	return pairs
}
func (m *OrderedMap[K, V]) Copy() *OrderedMap[K, V] {
	keys := make([]K, 0, m.Len())
	values := make([]V, 0, m.Len())
	indices := make(map[K]int, m.Len())
	var index int
	m.Each(func(key K, value V) {
		keys = append(keys, key)
		values = append(values, value)
		indices[key] = index
		index++
	})
	return &OrderedMap[K, V]{
		keys:           keys,
		values:         values,
		indices:        indices,
		deletedIndices: make(map[int]bool),
	}
}
func (m *OrderedMap[K, V]) Equals(other *OrderedMap[K, V]) bool {
	if m.Len() != other.Len() { return false }
	equals := true
	m.Each(func(key K, value V) {
		otherValue, ok := other.Get(key)
		if !ok || value != otherValue { equals = false }
	})
	return equals
}
