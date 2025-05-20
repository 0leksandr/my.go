package my

import (
	"reflect"
	"unsafe"
)

type ReflectStructGetter struct {
	rs reflect.Value
}
func (ReflectStructGetter) New(_struct any) ReflectStructGetter {
	checkIsStruct(_struct)

	rs := reflect.ValueOf(&_struct).Elem()
	if rs.Type().Kind() == reflect.Interface {
		val := reflect.ValueOf(_struct)
		addressableValue := reflect.New(val.Type()).Elem()
		addressableValue.Set(val)
		rs = addressableValue.Addr().Elem()
	}

	return ReflectStructGetter{rs}
}
func (reflectStruct ReflectStructGetter) NumField() int {
	return reflectStruct.rs.NumField()
}
func (reflectStruct ReflectStructGetter) Get(i int) any {
	rf := reflectStruct.rs.Field(i)

	return reflect.
		NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).
		Elem().
		Interface()
}

type ReflectStructSetter struct {
	rs reflect.Value
}
func (ReflectStructSetter) New(_struct any) ReflectStructSetter {
	checkIsStruct(_struct)
	return ReflectStructSetter{reflect.New(reflect.ValueOf(_struct).Type()).Elem()}
}
func (reflectStruct ReflectStructSetter) NumField() int {
	return reflectStruct.rs.NumField()
}
func (reflectStruct ReflectStructSetter) Set(i int, value any) {
	rf := reflectStruct.rs.Field(i)
	reflect.
		NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).
		Elem().
		Set(reflect.ValueOf(value))
}
func (reflectStruct ReflectStructSetter) Final() any {
	return reflectStruct.rs.Interface()
}

func GetExportedFields(_struct any) []any {
	checkIsStruct(_struct)
	var fields []any
	rs := reflect.ValueOf(_struct)
	for _, field := range reflect.VisibleFields(rs.Type()) {
		if field.IsExported() {
			fields = append(fields, rs.FieldByName(field.Name).Interface())
		}
	}
	return fields
}

func checkIsStruct(_struct any) {
	if reflect.ValueOf(_struct).Kind() != reflect.Struct {
		panic("not a struct")
	}
}
