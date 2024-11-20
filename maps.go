package my

import (
	"reflect"
	"unsafe"
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
			rs := reflect.ValueOf(&obj).Elem()
			if rs.Type().Kind() == reflect.Interface {
				addressableValue := reflect.New(val.Type()).Elem()
				addressableValue.Set(val)
				rs = addressableValue.Addr().Elem()
			}
			rs2 := reflect.New(val.Type()).Elem()
			for i := 0; i < val.NumField(); i++ {
				rf := rs.Field(i)
				rf2 := rs2.Field(i)
				reflect.
					NewAt(rf2.Type(), unsafe.Pointer(rf2.UnsafeAddr())).
					Elem().
					Set(reflect.ValueOf(InitMaps(reflect.NewAt(
						rf.Type(),
						unsafe.Pointer(rf.UnsafeAddr()),
					).Elem().Interface())))
			}
			return rs2.Interface().(T)
		//case reflect.Ptr:
		//	return reflect.ValueOf(InitMaps(val.Elem().Interface())).Addr().Interface().(T)
		default:
			return obj
	}
}
