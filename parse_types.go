package my

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	types2 "go/types"
	"io/fs"
	"path"
	"runtime"
)

type ParsedPackage struct {
	Structs    map[string]ParsedStruct
	Interfaces map[string]ParsedInterface
}

type ParsedInterface struct {
	Methods map[string]ParsedMethod
}

type ParsedStruct struct {
	Methods map[string]ParsedMethod
}

type ParsedMethod struct {
	Signature string
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
		conf := types2.Config{Importer: importer.Default()}
		files := make([]*ast.File, 0, len(astPkg.Files))
		for _, astFile := range astPkg.Files {
			files = append(files, astFile)
		}
		typesPkg, _ := conf.Check(dir, fset, files, nil)
		scope := typesPkg.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			t := obj.Type()
			if named, okNamed := t.(*types2.Named); okNamed {
				getMethods := func(hasMethods interface{
					NumMethods() int
					Method(int) *types2.Func
				}) map[string]ParsedMethod {
					methods := make(map[string]ParsedMethod)
					for i := 0; i < hasMethods.NumMethods(); i++ {
						method := hasMethods.Method(i)
						methods[method.Name()] = ParsedMethod{
							Signature: method.Type().String(), // MAYBE: trim argument names
							//Signature: named.Underlying().String(), // MAYBE: trim argument names
						}
					}
					return methods
				}
				underlying := named.Underlying()
				if _interface, ok := underlying.(*types2.Interface); ok {
					parsedInterfaces[named.Obj().Name()] = ParsedInterface{Methods: getMethods(_interface)}
				}
				if _, ok := underlying.(*types2.Struct); ok {
					parsedStructs[named.Obj().Name()] = ParsedStruct{Methods: getMethods(named)}
				}
			}
		}
	}

	return ParsedPackage{
		Structs:    parsedStructs,
		Interfaces: parsedInterfaces,
	}
}
