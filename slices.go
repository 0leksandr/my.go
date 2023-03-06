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
