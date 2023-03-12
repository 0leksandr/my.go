package my

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"reflect"
)

type ParsedPackage struct {
	structs    map[string]ParsedStruct
	interfaces map[string]ParsedInterface
}
func (parsedPackage ParsedPackage) Structs() map[string]ParsedStruct {
	return parsedPackage.structs
}
func (parsedPackage ParsedPackage) Interfaces() map[string]ParsedInterface {
	return parsedPackage.interfaces
}

type ParsedStruct struct {
	embedded []string
	methods  map[string]ParsedFuncType
}
func (parsedStruct ParsedStruct) Implements(parsedInterface ParsedInterface) bool { // MAYBE: remove
	for methodName, interfaceMethod := range parsedInterface.methods {
		if structMethod, ok := parsedStruct.methods[methodName]; ok {
			if !structMethod.SignatureEquals(interfaceMethod) { return false }
		} else {
			return false
		}
	}
	return true
}
func (parsedStruct ParsedStruct) Overrides(other ParsedStruct) bool { // MAYBE: rename
	ignoredMethods := []string{"New"}

	for methodName, otherMethod := range other.methods {
		if !InArray(methodName, ignoredMethods) {
			if thisMethod, ok := parsedStruct.methods[methodName]; ok {
				if !thisMethod.SignatureEquals(otherMethod) {
					return false
				}
			} else {
				return false
			}
		}
	}
	return true
}
func (parsedStruct ParsedStruct) ImplementsReal(_interface reflect.Type) bool {
	if _interface.Kind() != reflect.Interface { panic("must be an interface") }
	for i := 0; i < _interface.NumMethod(); i++ {
		if method := _interface.Method(i); method.IsExported() {
			if parsedMethod, ok := parsedStruct.methods[method.Name]; ok {
				if !parsedMethod.SignatureEqualsReal(method) { return false }
			} else {
				return false
			}
		}
	}
	return true
}

type ParsedFuncType struct {
	in  []ParsedType
	out []ParsedType
}
func (parsedFuncType ParsedFuncType) SignatureEquals(other ParsedFuncType) bool {
	if len(parsedFuncType.in) != len(other.in) { return false }
	if len(parsedFuncType.out) != len(other.out) { return false }
	for i, in := range parsedFuncType.in {
		if !in.Equals(other.in[i]) { return false }
	}
	for i, out := range parsedFuncType.out {
		if !out.Equals(other.out[i]) { return false }
	}
	return true
}
func (parsedFuncType ParsedFuncType) SignatureEqualsReal(method reflect.Method) bool {
	return parsedFuncType.EqualsReal(method.Type)
}
func (parsedFuncType ParsedFuncType) Equals(other ParsedType) bool {
	if otherParsedFuncType, ok := other.(ParsedFuncType); ok {
		return parsedFuncType.SignatureEquals(otherParsedFuncType)
	} else {
		return false
	}
}
func (parsedFuncType ParsedFuncType) EqualsReal(t reflect.Type) bool {
	if t.Kind() == reflect.Func {
		if len(parsedFuncType.in) != t.NumIn() { return false }
		if len(parsedFuncType.out) != t.NumOut() { return false }
		for i, in := range parsedFuncType.in {
			if !in.EqualsReal(t.In(i)) { return false }
		}
		for i, out := range parsedFuncType.out {
			if !out.EqualsReal(t.Out(i)) { return false }
		}
		return true
	} else {
		return false
	}
}

type ParsedType interface {
	Equals(ParsedType) bool
	EqualsReal(reflect.Type) bool
}

type ParsedInterface struct {
	methods map[string]ParsedFuncType
}
func (parsedInterface ParsedInterface) Equals(other ParsedType) bool {
	if otherParsedInterface, ok1 := other.(ParsedInterface); ok1 {
		if len(parsedInterface.methods) != len(otherParsedInterface.methods) { return false }
		for methodName, parsedMethod := range parsedInterface.methods {
			if otherParsedMethod, ok2 := otherParsedInterface.methods[methodName]; ok2 {
				if !parsedMethod.SignatureEquals(otherParsedMethod) { return false }
			} else {
				return false
			}
		}
		return true
	} else {
		return false
	}
}
func (parsedInterface ParsedInterface) EqualsReal(t reflect.Type) bool {
	if t.Kind() == reflect.Interface {
		if len(parsedInterface.methods) != t.NumMethod() { return false }
		for methodName, parsedMethod := range parsedInterface.methods {
			if otherMethod, ok := t.MethodByName(methodName); ok {
				if !parsedMethod.SignatureEqualsReal(otherMethod) { return false }
			} else {
				return false
			}
		}
		return true
	} else {
		return false
	}
}

type ParsedNamedType struct {
	literalName string
}
func (parsedNamedType ParsedNamedType) Equals(other ParsedType) bool {
	if otherParsedNamedType, ok := other.(ParsedNamedType); ok {
		return parsedNamedType.literalName == otherParsedNamedType.literalName
	} else {
		return false
	}
}
func (parsedNamedType ParsedNamedType) EqualsReal(t reflect.Type) bool {
	if parsedNamedType.literalName == "any" { parsedNamedType.literalName = "interface {}" }
	return t.String() == parsedNamedType.literalName
}

type ParsedArrayType struct {
	length      int
	elementType ParsedType
}
func (parsedArrayType ParsedArrayType) Equals(other ParsedType) bool {
	if otherParsedArrayType, ok := other.(ParsedArrayType); ok {
		return parsedArrayType.length == otherParsedArrayType.length &&
			parsedArrayType.elementType.Equals(otherParsedArrayType.elementType)
	} else {
		return false
	}
}
func (parsedArrayType ParsedArrayType) EqualsReal(t reflect.Type) bool {
	if t.Kind() == reflect.Array {
		return parsedArrayType.length == t.Len() &&
			parsedArrayType.elementType.EqualsReal(t.Elem())
	} else {
		return false
	}
}

type ParsedMapType struct {
	keyType     ParsedType
	elementType ParsedType
}
func (parsedMapType ParsedMapType) Equals(other ParsedType) bool {
	if otherParsedMapType, ok := other.(ParsedMapType); ok {
		return parsedMapType.keyType.Equals(otherParsedMapType.keyType) &&
			parsedMapType.elementType.Equals(otherParsedMapType.elementType)
	} else {
		return false
	}
}
func (parsedMapType ParsedMapType) EqualsReal(t reflect.Type) bool {
	if t.Kind() == reflect.Map {
		return parsedMapType.keyType.EqualsReal(t.Key()) &&
			parsedMapType.elementType.EqualsReal(t.Elem())
	} else {
		return false
	}
}

type ParsedChanType struct {
	valueType ParsedType
}
func (parsedChanType ParsedChanType) Equals(other ParsedType) bool {
	if otherParsedChanType, ok := other.(ParsedChanType); ok {
		return parsedChanType.valueType.Equals(otherParsedChanType.valueType)
	} else {
		return false
	}
}
func (parsedChanType ParsedChanType) EqualsReal(t reflect.Type) bool {
	if t.Kind() == reflect.Chan {
		return parsedChanType.valueType.EqualsReal(t.Elem())
	} else {
		return false
	}
}

type ParsedEllipsisType struct {
	elementType ParsedType
}
func (parsedEllipsisType ParsedEllipsisType) Equals(other ParsedType) bool {
	if otherParsedEllipsisType, ok := other.(ParsedEllipsisType); ok {
		return parsedEllipsisType.elementType.Equals(otherParsedEllipsisType.elementType)
	} else {
		return false
	}
}
func (parsedEllipsisType ParsedEllipsisType) EqualsReal(reflect.Type) bool {
	return false // TODO: fix
}

func ParseTypes() ParsedPackage {
	return parseTypes(path.Dir(GetTrace(true).SkipFile(1)[0].File))
}
func parseTypes(dir string) ParsedPackage {
	//_, errImport := importer.Default().Import(dir)
	//PanicIf(errImport)

	parsedStructs := make(map[string]ParsedStruct)
	parsedInterfaces := make(map[string]ParsedInterface)

	fset := token.NewFileSet()
	pkgs, errParseDir := parser.ParseDir(
		fset,
		dir,
		func(fs.FileInfo) bool { return true },
		parser.AllErrors,
	)
	PanicIf(errParseDir)
	for _, astPkg := range pkgs {
		//conf := types2.Config{Importer: importer.Default()}
		//files := make([]*ast.File, 0, len(astPkg.Files))
		//for _, astFile := range astPkg.Files {
		//	files = append(files, astFile)
		//}
		//typesPkg, errCheck := conf.Check(dir, fset, files, nil)
		//PanicIf(errCheck)
		//scope := typesPkg.Scope()
		//for _, name := range scope.Names() {
		//	obj := scope.Lookup(name)
		//	t := obj.Type()
		//	if named, okNamed := t.(*types2.Named); okNamed {
		//		getMethods := func(hasMethods interface{
		//			NumMethods() int
		//			Method(int) *types2.Func
		//		}) map[string]ParsedFuncType {
		//			methods := make(map[string]ParsedFuncType)
		//			for i := 0; i < hasMethods.NumMethods(); i++ {
		//				method := hasMethods.Method(i)
		//				methods[method.Name()] = ParsedFuncType{
		//					Signature: method.Type().String(), // MAYBE: trim argument names
		//				}
		//			}
		//			return methods
		//		}
		//		underlying := named.Underlying()
		//		if _interface, ok := underlying.(*types2.Interface); ok {
		//			parsedInterfaces[named.Obj().Name()] = ParsedInterface{methods: getMethods(_interface)}
		//		}
		//		if _, ok := underlying.(*types2.Struct); ok {
		//			parsedStructs[named.Obj().Name()] = ParsedStruct{methods: getMethods(named)}
		//		}
		//	}
		//}

		walkDecls := func(f func(ast.Decl)) {
			for _, astFile := range astPkg.Files {
				for _, decl := range astFile.Decls {
					f(decl)
				}
			}
		}
		walkDecls(func(decl ast.Decl) {
			if genDecl, isGenDecl := decl.(*ast.GenDecl); isGenDecl {
				for _, spec := range genDecl.Specs {
					if typeSpec, isTypeSpec := spec.(*ast.TypeSpec); isTypeSpec {
						specName := typeSpec.Name.Name
						astType := typeSpec.Type
						if astStructType, isStructType := astType.(*ast.StructType); isStructType {
							parsedStruct := ParsedStruct{methods: make(map[string]ParsedFuncType)}
							for _, field := range astStructType.Fields.List {
								if field.Names == nil {
									if named, ok := parseType(field.Type).(ParsedNamedType); ok {
										parsedStruct.embedded = append(parsedStruct.embedded, named.literalName)
									}
								}
							}
							parsedStructs[specName] = parsedStruct
						} else if astInterfaceType, isInterfaceType := astType.(*ast.InterfaceType); isInterfaceType {
							parsedInterfaces[specName] = parseInterface(astInterfaceType)
						}
					}
				}
			}
		})
		walkDecls(func(decl ast.Decl) {
			if funcDecl, isFuncDecl := decl.(*ast.FuncDecl); isFuncDecl {
				if receivers := funcDecl.Recv; receivers != nil {
					if len(receivers.List) == 1 {
						receiverAstField := receivers.List[0]
						receiverType := receiverAstField.Type
						if starExpr, isStarExpr := receiverType.(*ast.StarExpr); isStarExpr {
							receiverType = starExpr.X
						}
						if ident, isIdent := receiverType.(*ast.Ident); isIdent {
							receiverName := ident.Name
							if _, structExists := parsedStructs[receiverName]; structExists {
								parsedStructs[receiverName].methods[funcDecl.Name.Name] = parseFunc(funcDecl.Type)
							}
						}
					}
				}
			}
		})
	}

	return ParsedPackage{
		structs:    parsedStructs,
		interfaces: parsedInterfaces,
	}
}
func parseInterface(astInterfaceType *ast.InterfaceType) ParsedInterface {
	parsedMethods := make(map[string]ParsedFuncType)
	for _, methodAstField := range astInterfaceType.Methods.List {
		if astFuncType, isFuncType := methodAstField.Type.(*ast.FuncType); isFuncType {
			names := methodAstField.Names
			if len(names) != 1 {
				Dump2(astInterfaceType)
				panic("non-singular method name")
			}
			parsedMethods[names[0].Name] = parseFunc(astFuncType)
		}
	}

	return ParsedInterface{methods: parsedMethods}
}
func parseFunc(astFuncType *ast.FuncType) ParsedFuncType {
	var in []ParsedType
	paramsList := astFuncType.Params.List
	if len(paramsList) > 0 {
		in = make([]ParsedType, 0, len(paramsList))
		for _, astField := range paramsList {
			in = append(in, parseType(astField.Type))
		}
	}

	var out []ParsedType
	if results := astFuncType.Results; results != nil {
		resultsList := results.List
		if len(resultsList) > 0 {
			out = make([]ParsedType, 0, len(resultsList))
			for _, astField := range resultsList {
				out = append(out, parseType(astField.Type))
			}
		}
	}

	return ParsedFuncType{
		in:  in,
		out: out,
	}
}
func parseType(astExpr ast.Expr) ParsedType {
	switch astExpr.(type) {
		case *ast.Ident:
			return ParsedNamedType{astExpr.(*ast.Ident).Name}
		case *ast.InterfaceType:
			return parseInterface(astExpr.(*ast.InterfaceType))
		case *ast.SelectorExpr:
			return ParsedNamedType{fmt.Sprintf(
				"%s.%s",
				astExpr.(*ast.SelectorExpr).X.(*ast.Ident).Name,
				astExpr.(*ast.SelectorExpr).Sel.Name,
			)}
		case *ast.StarExpr:
			parsedX := parseType(astExpr.(*ast.StarExpr).X)
			if parsedNamedType, isNamedType := parsedX.(ParsedNamedType); isNamedType {
				return ParsedNamedType{"*" + parsedNamedType.literalName}
			} else {
				panic("ast.StarExpr.X is not a named type")
			}
		case *ast.ArrayType:
			if astExpr.(*ast.ArrayType).Len != nil {
				panic("TODO")
			}
			return ParsedArrayType{
				length:      -1,
				elementType: parseType(astExpr.(*ast.ArrayType).Elt),
			}
		case *ast.MapType:
			return ParsedMapType{
				keyType:     parseType(astExpr.(*ast.MapType).Key),
				elementType: parseType(astExpr.(*ast.MapType).Value),
			}
		case *ast.ChanType:
			return ParsedChanType{
				valueType: parseType(astExpr.(*ast.ChanType).Value),
			}
		case *ast.Ellipsis:
			return ParsedEllipsisType{
				elementType: parseType(astExpr.(*ast.Ellipsis).Elt),
			}
		case *ast.FuncType:
			return parseFunc(astExpr.(*ast.FuncType))
		default:
			Dump2(astExpr)
			panic("cannot parse expr as type")
	}
}
