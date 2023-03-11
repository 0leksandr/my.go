package my

func Revert[V any](slice []V) []V {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
}
func Remove[V any](slice []V, nth int) []V {
	last := len(slice)-1
	if nth != last { slice[nth] = slice[last] }
	return slice[:last]
}
func InArray[V comparable](needle V, haystack []V) bool {
	for _, element := range haystack {
		if element == needle {
			return true
		}
	}
	return false
}
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for key := range m { keys = append(keys, key) }
	return keys
}
func ArrayFilter[V any](a []V, f func(V) bool) []V {
	var filtered []V
	for _, v := range a {
		if f(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}
func ArrayMap[T1 any, T2 any](a []T1, f func(T1) T2) []T2 {
	mapped := make([]T2, len(a))
	for i, t1 := range a { mapped[i] = f(t1) }
	return mapped
}
