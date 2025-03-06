package my

import (
	"reflect"
)

// InitMaps replaces all nil-maps with initialized zero-length maps
func InitMaps[T any](obj T) T {
	val := reflect.ValueOf(obj)

	switch val.Kind() {
		case reflect.Map:
			if val.IsNil() {
				return reflect.MakeMap(val.Type()).Interface().(T)
			}
			newMap := reflect.MakeMap(val.Type())
			for _, k := range val.MapKeys() {
				newMap.SetMapIndex(k, reflect.ValueOf(InitMaps(val.MapIndex(k).Interface())))
			}
			return newMap.Interface().(T)
		case reflect.Struct:
			getter := ReflectStructGetter{}.New(obj)
			setter := ReflectStructSetter{}.New(obj)
			for i := 0; i < getter.NumField(); i++ {
				setter.Set(i, InitMaps(getter.Get(i)))
			}
			return setter.Final().(T)
		//case reflect.Ptr:
		//	return reflect.ValueOf(InitMaps(val.Elem().Interface())).Addr().Interface().(T)
		default:
			return obj
	}
}
