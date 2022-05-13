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
func (m *OrderedMap) Add(key interface{}, value interface{}) {
	if m.Has(key) { panic("key already filled") }
	m.keys = append(m.keys, key)
	m.values = append(m.values, value)
	m.indices[key] = len(m.keys) - 1
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
func (m OrderedMap) Pairs() []OrderedMapPair {
	pairs := make([]OrderedMapPair, 0, m.Len())
	for index, key := range m.keys {
		if !m.deletedIndices[index] {
			pairs = append(
				pairs,
				OrderedMapPair{
					Key:   key,
					Value: m.values[index],
				},
			)
		}
	}
	return pairs
}
