package my

import (
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"unsafe"
)

//go:linkname reflect_typelinks reflect.typelinks
func reflect_typelinks() ([]unsafe.Pointer, [][]int32)

//go:linkname reflect_add reflect.add
func reflect_add(unsafe.Pointer, uintptr, string) unsafe.Pointer

func Types(all bool) []reflect.Type {
	var prefix string
	if !all { prefix = pkgName(Trace{}.New().SkipFile(1)[0].File) + "." }
	return getTypes(prefix)
}
func getTypes(prefix string) []reflect.Type {
	types := make([]reflect.Type, 0)
	sections, offsets := reflect_typelinks()
	for i, base := range sections {
		for _, offset := range offsets[i] {
			typeAddr := reflect_add(base, uintptr(offset), "I am an idiot")
			t := reflect.TypeOf(*(*interface{})(unsafe.Pointer(&typeAddr)))
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
				if strings.HasPrefix(t.String(), prefix) {
					if !strings.ContainsAny(t.String(), "[]") {
						types = append(types, t)
					}
				}
			}
		}
	}

	return types
}
func pkgName(filename string) string {
	file, err := parser.ParseFile(token.NewFileSet(), filename, nil, parser.PackageClauseOnly)
	PanicIf(err)
	return file.Name.Name
}
