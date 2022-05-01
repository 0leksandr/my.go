package my

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"runtime"
)

type ParsedPackage struct {
	Structs    map[string]ParsedStruct
	Interfaces map[string]ParsedInterface
}

type ParsedStruct struct {
	Methods map[string]ParsedFunc
}
func (parsedStruct ParsedStruct) Implements(parsedInterface ParsedInterface) bool {
	for methodName, interfaceMethod := range parsedInterface.Methods {
		if structMethod, ok := parsedStruct.Methods[methodName]; ok {
			if !structMethod.SignatureEquals(interfaceMethod) { return false }
		} else {
			return false
		}
	}
	return true
}

type ParsedFunc struct {
	In  []ParsedType
	Out []ParsedType
}
func (parsedFunc ParsedFunc) SignatureEquals(other ParsedFunc) bool {
	if len(parsedFunc.In) != len(other.In) { return false }
	if len(parsedFunc.Out) != len(other.Out) { return false }
	for i, in := range parsedFunc.In {
		if !in.Equals(other.In[i]) { return false }
	}
	for i, out := range parsedFunc.Out {
		if !out.Equals(other.Out[i]) { return false }
	}
	return true
}

type ParsedType interface {
	Equals(other ParsedType) bool
}

type ParsedInterface struct {
	Methods map[string]ParsedFunc
}
func (parsedInterface ParsedInterface) Equals(other ParsedType) bool {
	if otherParsedInterface, ok1 := other.(ParsedInterface); ok1 {
		if len(parsedInterface.Methods) != len(otherParsedInterface.Methods) { return false }
		for methodName, parsedMethod := range parsedInterface.Methods {
			if otherParsedMethod, ok2 := otherParsedInterface.Methods[methodName]; ok2 {
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

type ParsedNamedType struct {
	LiteralName string
}
func (parsedNamedType ParsedNamedType) Equals(other ParsedType) bool {
	if otherParsedNamedType, ok := other.(ParsedNamedType); ok {
		return parsedNamedType.LiteralName == otherParsedNamedType.LiteralName
	} else {
		return false
	}
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

func ParseTypes() ParsedPackage {
	_, file, _, okCaller := runtime.Caller(1)
	if !okCaller { panic(nil) }
	dir := path.Dir(file)
	if false { panic(dir) }

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
		//		}) map[string]ParsedFunc {
		//			methods := make(map[string]ParsedFunc)
		//			for i := 0; i < hasMethods.NumMethods(); i++ {
		//				method := hasMethods.Method(i)
		//				methods[method.Name()] = ParsedFunc{
		//					Signature: method.Type().String(), // MAYBE: trim argument names
		//				}
		//			}
		//			return methods
		//		}
		//		underlying := named.Underlying()
		//		if _interface, ok := underlying.(*types2.Interface); ok {
		//			parsedInterfaces[named.Obj().Name()] = ParsedInterface{Methods: getMethods(_interface)}
		//		}
		//		if _, ok := underlying.(*types2.Struct); ok {
		//			parsedStructs[named.Obj().Name()] = ParsedStruct{Methods: getMethods(named)}
		//		}
		//	}
		//}

		for _, astFile := range astPkg.Files {
			for _, decl := range astFile.Decls {
				if genDecl, isGenDecl := decl.(*ast.GenDecl); isGenDecl {
					for _, spec := range genDecl.Specs {
						if typeSpec, isTypeSpec := spec.(*ast.TypeSpec); isTypeSpec {
							specName := typeSpec.Name.Name
							astType := typeSpec.Type
							if _, isStructType := astType.(*ast.StructType); isStructType {
								parsedStructs[specName] = ParsedStruct{Methods: make(map[string]ParsedFunc)}
							}
							if astInterfaceType, isInterfaceType := astType.(*ast.InterfaceType); isInterfaceType {
								parsedInterfaces[specName] = parseInterface(astInterfaceType)
							}
						}
					}
				}
			}
		}
		for _, astFile := range astPkg.Files {
			for _, decl := range astFile.Decls {

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
									parsedStructs[receiverName].Methods[funcDecl.Name.Name] = parseFunc(funcDecl.Type)
								}
							}
						}
					}
				}
			}
		}
	}

	return ParsedPackage{
		Structs:    parsedStructs,
		Interfaces: parsedInterfaces,
	}
}
func parseInterface(astInterfaceType *ast.InterfaceType) ParsedInterface {
	parsedMethods := make(map[string]ParsedFunc)
	for _, methodAstField := range astInterfaceType.Methods.List {
		names := methodAstField.Names
		if len(names) != 1 { panic("non-singular method name") }
		methodName := names[0].Name

		_type := methodAstField.Type
		if astFuncType, isFuncType := _type.(*ast.FuncType); isFuncType {
			parsedMethods[methodName] = parseFunc(astFuncType)
		} else {
			panic("method is not a func")
		}
	}

	return ParsedInterface{Methods: parsedMethods}
}
func parseFunc(astFuncType *ast.FuncType) ParsedFunc {
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

	return ParsedFunc{
		In:  in,
		Out: out,
	}
}
func parseType(astExpr ast.Expr) ParsedType {
	if astIdent, isIdent := astExpr.(*ast.Ident); isIdent {
		return ParsedNamedType{astIdent.Name}
	} else if astInterfaceType, isInterface := astExpr.(*ast.InterfaceType); isInterface {
		return parseInterface(astInterfaceType)
	} else if astSelectorExpr, isSelectorExpr := astExpr.(*ast.SelectorExpr); isSelectorExpr {
		return ParsedNamedType{fmt.Sprintf(
			"%s.%s",
			astSelectorExpr.X.(*ast.Ident).Name,
			astSelectorExpr.Sel.Name,
		)}
	} else if astStarExpr, isStarExpr := astExpr.(*ast.StarExpr); isStarExpr {
		parsedX := parseType(astStarExpr.X)
		if parsedNamedType, isNamedType := parsedX.(ParsedNamedType); isNamedType {
			return ParsedNamedType{"*" + parsedNamedType.LiteralName}
		} else {
			panic("ast.StarExpr.X is not a named type")
		}
	} else if astArrayType, isArrayType := astExpr.(*ast.ArrayType); isArrayType {
		if astArrayType.Len != nil { panic("TODO") }
		return ParsedArrayType{
			length:      -1,
			elementType: parseType(astArrayType.Elt),
		}
	} else {
		Dump2(astExpr)
		panic("cannot parse expr as type")
	}
}
