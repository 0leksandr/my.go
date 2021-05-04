package my

type OrderedMap struct { // TODO: remove
	keys    []interface{}
	values  []interface{}
	indices map[interface{}]int
}
type OrderedMapPair struct {
	key   interface{}
	value interface{}
}
func NewOrderedMap() OrderedMap {
	return OrderedMap{
		keys:    make([]interface{}, 0),
		values:  make([]interface{}, 0),
		indices: make(map[interface{}]int),
	}
}
func (m *OrderedMap) Set(key interface{}, value interface{}) {
	if index, ok := m.indices[key]; ok {
		m.values[index] = value
	} else {
		m.keys = append(m.keys, key)
		m.values = append(m.values, value)
		m.indices[key] = m.Len() - 1
	}
}
func (m OrderedMap) Len() int {
	return len(m.keys)
}
func (m OrderedMap) Get(key interface{}) interface{} {
	if index, ok := m.indices[key]; ok {
		return m.values[index]
	}
	panic("key not found")
}
func (m OrderedMap) Has(key interface{}) bool {
	_, ok := m.indices[key]
	return ok
}
func (m OrderedMap) Keys() []interface{} {
	return m.keys
}
func (m OrderedMap) Values() []interface{} {
	return m.values
}
func (m OrderedMap) Pairs() []OrderedMapPair {
	pairs := make([]OrderedMapPair, 0, m.Len())
	for index, key := range m.keys {
		pairs = append(
			pairs,
			OrderedMapPair{
				key:   key,
				value: m.values[index],
			},
		)
	}
	return pairs
}
