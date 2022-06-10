package my

type OrderedMap struct {
	keys           []interface{}
	values         []interface{}
	indices        map[interface{}]int
	deletedIndices map[int]bool
}
type OrderedMapPair struct {
	Key   interface{}
	Value interface{}
}
func (OrderedMap) New() OrderedMap {
	return OrderedMap{
		keys:           make([]interface{}, 0),
		values:         make([]interface{}, 0),
		indices:        make(map[interface{}]int),
		deletedIndices: make(map[int]bool),
	}
}
func (m *OrderedMap) Add(key interface{}, value interface{}) *OrderedMap {
	if m.Has(key) { panic("key already filled") }
	m.keys = append(m.keys, key)
	m.values = append(m.values, value)
	m.indices[key] = len(m.keys) - 1
	return m
}
func (m OrderedMap) Len() int {
	return len(m.indices)
}
func (m OrderedMap) Get(key interface{}) (interface{}, bool) {
	if index, ok := m.indices[key]; ok {
		return m.values[index], true
	}
	return nil, false
}
func (m OrderedMap) Has(key interface{}) bool {
	_, ok := m.indices[key]
	return ok
}
func (m *OrderedMap) Del(key interface{}) {
	if index, ok := m.indices[key]; ok {
		m.deletedIndices[index] = true
		delete(m.indices, m.keys[index])

		if len(m.deletedIndices) > len(m.indices) {
			l := len(m.indices) + len(m.deletedIndices)
			indices := make([]int, l)
			for i := 0; i < l; i++ { indices[i] = i }
			for _index := range m.deletedIndices { indices[_index] = -1 }

			keys := make([]interface{}, 0, len(m.indices))
			values := make([]interface{}, 0, len(m.indices))
			for _, _index := range indices {
				if _index != -1 {
					keys = append(keys, m.keys[_index])
					values = append(values, m.values[_index])
				}
			}

			m.keys = keys
			m.values = values
			m.deletedIndices = make(map[int]bool)
		}
	}
}
func (m OrderedMap) Each(f func(key interface{}, value interface{})) {
	for index, key := range m.keys {
		if !m.deletedIndices[index] {
			f(key, m.values[index])
		}
	}
}
func (m OrderedMap) Pairs() []OrderedMapPair {
	pairs := make([]OrderedMapPair, 0, m.Len())
	m.Each(func(key interface{}, value interface{}) {
		pairs = append(
			pairs,
			OrderedMapPair{
				Key:   key,
				Value: value,
			},
		)
	})
	return pairs
}
func (m OrderedMap) Copy() OrderedMap {
	keys := make([]interface{}, 0, m.Len())
	values := make([]interface{}, 0, m.Len())
	indices := make(map[interface{}]int, m.Len())
	var index int
	m.Each(func(key interface{}, value interface{}) {
		keys = append(keys, key)
		values = append(values, value)
		indices[key] = index
		index++
	})
	return OrderedMap{
		keys:           keys,
		values:         values,
		indices:        indices,
		deletedIndices: make(map[int]bool),
	}
}
func (m OrderedMap) Equals(other OrderedMap) bool {
	if m.Len() != other.Len() { return false }
	equals := true
	m.Each(func(key interface{}, value interface{}) {
		otherValue, ok := other.Get(key)
		if !ok || value != otherValue { equals = false }
	})
	return equals
}
