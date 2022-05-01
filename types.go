package my

import (
	"go/parser"
	"go/token"
	"reflect"
	"runtime"
	"unsafe"
)

//go:linkname reflect_typelinks reflect.typelinks
func reflect_typelinks() ([]unsafe.Pointer, [][]int32)

//go:linkname reflect_add reflect.add
func reflect_add(unsafe.Pointer, uintptr, string) unsafe.Pointer

func Types(all bool) []reflect.Type {
	types := make([]reflect.Type, 0)
	var pkgPath string
	if !all { pkgPath = pkgName(1) + "." }
	sections, offsets := reflect_typelinks()
	for i, base := range sections {
		for _, offset := range offsets[i] {
			typeAddr := reflect_add(base, uintptr(offset), "I'm an idiot")
			t := reflect.TypeOf(*(*interface{})(unsafe.Pointer(&typeAddr)))
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
				if pkgPath == "" || stringStartsWith(t.String(), pkgPath) {
					types = append(types, t)
				}
			}
		}
	}

	return types
}
func pkgName(skip int) string {
	_, filename, _, ok := runtime.Caller(skip + 1)
	if !ok { panic("could not detect filename") }
	file, err := parser.ParseFile(token.NewFileSet(), filename, nil, parser.PackageClauseOnly)
	PanicIf(err)
	return file.Name.Name
}
func stringStartsWith(str string, prefix string) bool {
	if len(str) < len(prefix) { return false }
	return str[:len(prefix)] == prefix
}
