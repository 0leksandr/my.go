package my

import "reflect"

func Revert(slice interface{}) interface{} {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice { panic("argument must be a slice") }
	for i, j := 0, v.Len()-1; i < j; i, j = i+1, j-1 {
		vi, vj := v.Index(i).Interface(), v.Index(j).Interface()
		v.Index(i).Set(reflect.ValueOf(vj))
		v.Index(j).Set(reflect.ValueOf(vi))
	}
	return v.Interface()
}
func Remove(slice interface{}, nth int) interface{} {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice { panic("argument must be a slice") }
	last := v.Len()-1
	if nth != last { v.Index(nth).Set(v.Index(last)) }
	return v.Slice(0, last).Interface()
}
func InArray(needle interface{}, haystack interface{}) bool {
	vHaystack := reflect.ValueOf(haystack)
	if vHaystack.Kind() != reflect.Slice { panic("argument must be a slice") }
	if reflect.TypeOf(needle) != vHaystack.Type().Elem() { panic("types mismatch") }
	for i, l := 0, vHaystack.Len(); i < l; i++ {
		if vHaystack.Index(i).Interface() == needle { return true }
	}
	return false
}
