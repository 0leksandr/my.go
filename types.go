package my

import (
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"unsafe"
)

//go:linkname reflect_typelinks reflect.typelinks
//goland:noinspection GoSnakeCaseUsage
func reflect_typelinks() ([]unsafe.Pointer, [][]int32)

//go:linkname reflect_add reflect.add
//goland:noinspection GoSnakeCaseUsage,GoUnusedParameter
func reflect_add(unsafe.Pointer, uintptr, string) unsafe.Pointer

func Types(all bool) []reflect.Type {
	var prefix string
	if !all { prefix = pkgName(Trace{}.New().SkipFile(1)[0].File) + "." }
	return getTypes(prefix)
}
func getTypes(prefix string) []reflect.Type {
	var types []reflect.Type
	type TypeOffset struct {
		t      reflect.Type
		offset int32
	}
	sections, offsets := reflect_typelinks()
	for i, base := range sections {
		var typeOffsets []TypeOffset
		for _, offset := range offsets[i] {
			typeAddr := reflect_add(base, uintptr(offset), "I am an idiot")
			t := reflect.TypeOf(*(*interface{})(unsafe.Pointer(&typeAddr)))
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
				if strings.HasPrefix(t.String(), prefix) {
					if !strings.Contains(t.String(), "·") {
						typeOffsets = append(typeOffsets, TypeOffset{t, offset})
					}
				}
			}
		}

		minOffset := (func() int32 {
			if len(typeOffsets) == 0 {
				return 0
			} else {
				minOffset := typeOffsets[0].offset
				for _, typeOffset := range typeOffsets[1:] {
					if typeOffset.offset < minOffset {
						minOffset = typeOffset.offset
					}
				}
				return minOffset
			}
		})()
		minOffset += 10000 // TODO: make it normal
		for _, typeOffset := range typeOffsets {
			if typeOffset.t.Kind() == reflect.Interface || typeOffset.offset >= minOffset {
				types = append(types, typeOffset.t)
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
