package my

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/packages"
	"os"
	"path"
	"reflect"
	"regexp"
)

func OnType[T any](root any, fn func(_ T, parents []*any) bool) {
	nonNullableTypeKinds := map[reflect.Kind]struct{}{}
	isNil := func(value reflect.Value) (result bool) {
		kind := value.Type().Kind()
		defer func() {
			if recover() != nil {
				nonNullableTypeKinds[kind] = struct{}{}
				result = false
			}
		}()
		if _, ok := nonNullableTypeKinds[kind]; ok {
			return false
		}
		return value.IsNil()
	}

	var processValue func(any, []*any)
	//typeProcessors := map[reflect.Type]func(any, []*any){}
	////typeProcessors := map[ast.TypeSpec]func(any, []any){}
	addresses := map[uintptr]struct{}{}
	processValue = func(value any, parents []*any) {
		if len(parents) > 10000 { panic("we went too deep") }

		if t, isT := value.(T); isT {
			if !fn(t, parents) {
				return
			}
		}

		if value == nil { return }
		rv := reflect.ValueOf(value)
		if isNil(rv) { return }
		processValue2 := func(value2 any) {
			processValue(value2, append(parents, &value))
		}
		switch rv.Type().Kind() {
			case reflect.Array, reflect.Slice:
				for i := 0; i < rv.Len(); i++ {
					processValue2(rv.Index(i).Interface())
				}
			case reflect.Interface:
				processValue2(rv.Elem().Interface())
			case reflect.Map:
				for _, key := range rv.MapKeys() {
					processValue2(key.Interface())
					processValue2(rv.MapIndex(key).Interface())
				}
			case reflect.Pointer:
				addr := rv.Pointer()
				if _, ok := addresses[addr]; ok {
					return
				} else {
					addresses[addr] = struct{}{}
				}

				processValue2(rv.Elem().Interface())
			case reflect.Struct:
				for _, field := range GetExportedFields(value) {
					processValue2(field)
				}
			//case reflect.Func:
			default:
		}
	}

	processValue(root, nil)
}

func parseGo(fn func(decl ast.Decl)) {
	parseGoDir(path.Dir(GetTrace(true).SkipFile(1)[0].File), fn)
}
func parseGoDir(dir string, fn func(ast.Decl)) {
	//fset := token.NewFileSet()
	//pkgs := MustFirst(parser.ParseDir(
	//	fset,
	//	dir,
	//	func(fs.FileInfo) bool { return true },
	//	parser.AllErrors,
	//))
	//for _, astPkg := range pkgs {
	//	//for _, astFile := range astPkg.Files {
	//	//}
	//
	//	config := types.Config{Importer: nil} // use the default importer
	//	//config := types.Config{Importer: importer.Default()}
	//	info := &types.Info{
	//		Types: make(map[ast.Expr]types.TypeAndValue),
	//		//Scopes: make(map[ast.Node]*types.Scope),
	//	}
	//	files := make([]*ast.File, 0, len(astPkg.Files))
	//	for _, file := range astPkg.Files { files = append(files, file) }
	//	MustFirst(config.Check(astPkg.Name, fset, files, info))
	//	for expr, tv := range info.Types {
	//		fmt.Printf("Expression: %v, Type: %v\n", expr, tv.Type.String())
	//		//fmt.Printf("Expression: %v, Type: %v\n", src[expr.Pos()-1:expr.End()-1], tv.Type.String())
	//	}
	//	Dump2(info)
	//	//Dump3(info)
	//}

	cfg := &packages.Config{
		Mode:  packages.LoadSyntax,
		Dir:   dir,
		Tests: false, // Skip test files (optional)
	}
	for _, pkg := range MustFirst(packages.Load(cfg, "./...")) {
		////config := types.Config{
		////	//Importer: &NullImporter{}, // Use the custom importer to skip external imports
		////	Importer: NewStrictModuleImporter(),
		////}
		////info := &types.Info{
		////	Types: make(map[ast.Expr]types.TypeAndValue),
		////}
		////MustFirst(config.Check(pkg.PkgPath, pkg.Fset, pkg.Syntax, info))
		////for expr, tv := range info.Types {
		////	fmt.Printf("Expression: %v, Type: %v\n", expr, tv.Type)
		////}
		//
		//for fileIdx, file := range pkg.Syntax {
		//	fileContent := MustFirst(os.ReadFile(pkg.CompiledGoFiles[fileIdx]))
		//	ast.Inspect(file, func(n ast.Node) bool {
        //        if expr, isExpr := n.(ast.Expr); isExpr {
        //            if tv, hasType := pkg.TypesInfo.Types[expr]; hasType {
        //                start := pkg.Fset.Position(expr.Pos())
        //                end := pkg.Fset.Position(expr.End())
        //                exprText := string(fileContent[start.Offset:end.Offset])
		//
        //                //fmt.Printf("Expression: %q\n", exprText)
        //                //fmt.Printf("Type: %v\n\n", tv.Type)
		//				//if regexp.MustCompile(".ItemRepository$").MatchString(tv.Type.String()) {
		//				//}
		//				if tv.Type.String() == "*github.com/MythicalGames/platform-items/internal/repository/postgresdb.ItemRepository" {
		//					Dump3(exprText)
		//				}
        //            }
        //        }
        //        return true
        //    })
		//}

		type MethodCall struct {
			//receiver ast.Expr
			method   string
			location string
		}
		methodCalls := make(map[string][]MethodCall, 0)
		dirRe := regexp.MustCompile("^" + regexp.QuoteMeta(dir) + "/")
		strip := func(s string) string { return dirRe.ReplaceAllString(s, "") }
		fileContents := map[string][]byte{}
		for fileIdx := range pkg.Syntax {
			fileContents[pkg.CompiledGoFiles[fileIdx]] = MustFirst(os.ReadFile(pkg.CompiledGoFiles[fileIdx]))
		}
		apiGenRe := regexp.MustCompile("^" + regexp.QuoteMeta(dir) + "/api/gen/")
		OnType(pkg.Syntax, func(callExpr ast.CallExpr, parents []*any) bool {
			if selectorExpr, isSelectorExpr := callExpr.Fun.(*ast.SelectorExpr); isSelectorExpr {
				p := pkg.Fset.Position(selectorExpr.Sel.Pos())
				if apiGenRe.MatchString(p.Filename) { return false } // TODO: refactor
				location := fmt.Sprintf("%s:%d:%d", strip(p.Filename), p.Line, p.Column)
				if tv, ok := pkg.TypesInfo.Types[selectorExpr.X]; ok {
					Dump3(pkg.TypesInfo.Types[callExpr.Fun].Type.String())
					typeName := tv.Type.String()
					if _, not := methodCalls[typeName]; !not {
						methodCalls[typeName] = []MethodCall{}
					}
					methodCalls[typeName] = append(
						methodCalls[typeName],
						MethodCall{
							method:   selectorExpr.Sel.Name,
							location: location,
						},
					)
				} else {
					//Dump3(location)
					//start := pkg.Fset.Position(callExpr.Pos())
					//end := pkg.Fset.Position(callExpr.End())
					//Dump3(string(fileContents[p.Filename][start.Offset:end.Offset]))
				}
			}
			return true
		})
		Dump3(pkg.Name)
		Dump3(methodCalls["*github.com/MythicalGames/platform-items/internal/repository/postgresdb.ItemRepository"])

		//calls := make(map[string]int)
		//OnType(pkg.Syntax, func(callExpr ast.CallExpr, parents []*any) bool {
		//	// what is `reflect.ValueOf(..).Kind()`?
		//	// How to know that `reflect.ValueOf` returns `reflect.Value`? There is nothing about it in `ast` call stack 😡
		//	// Another question: even if we are ignoring objects, methods and functions from external packages, how do
		//	// we have complete context of current symbol/call/stack?
		//	// Even if I parse contexts of functions, how do I know a value `a` is present _after_ some point and not
		//	// _before_ another?!
		//
		//	var typeOf func(ast.Expr) string
		//	typeOf = func(expr ast.Expr) string {
		//		if false {
		//			panic(nil)
		//		} else if ident, isIdent := expr.(*ast.Ident); isIdent {
		//			if ident.Obj != nil { return "" }
		//			return ident.Name
		//		} else if selectorExpr, isSelectorExpr := expr.(*ast.SelectorExpr); isSelectorExpr {
		//			typeOfX := typeOf(selectorExpr.X)
		//			if typeOfX == "" { return "" }
		//			return typeOfX + "." + selectorExpr.Sel.Name
		//		} else if _, isFuncLit := expr.(*ast.FuncLit); isFuncLit {
		//			// func(){}
		//			return ""
		//		} else {
		//			panic(expr)
		//		}
		//	}
		//
		//	var nameFromExpr func(expr ast.Expr) string
		//	nameFromExpr = func(expr ast.Expr) string {
		//		if true { return typeOf(expr) }
		//		if false {
		//			panic(nil)
		//		} else if ident, isIdent := expr.(*ast.Ident); isIdent {
		//			return ident.Name
		//		} else if selectorExpr, isSelectorExpr := expr.(*ast.SelectorExpr); isSelectorExpr {
		//			//return nameFromExpr(selectorExpr.X) + "." + selectorExpr.Sel.Name
		//			return typeOf(selectorExpr.X) + "." + selectorExpr.Sel.Name
		//		} else if _, isFuncLit := expr.(*ast.FuncLit); isFuncLit {
		//			// func(){}
		//			return ""
		//		} else if true {
		//			panic(expr)
		//		} else if callExpr, isCallExpr := expr.(*ast.CallExpr); isCallExpr {
		//			return nameFromExpr(callExpr.Fun)
		//		} else if compositeLit, isCompositeLit := expr.(*ast.CompositeLit); isCompositeLit {
		//			return nameFromExpr(compositeLit.Type) + "{}"
		//		} else if parenExpr, isParenExpr := expr.(*ast.ParenExpr); isParenExpr {
		//			return ""
		//			return "(" + nameFromExpr(parenExpr.X) + ")"
		//		} else if starExpr, isStarExpr := expr.(*ast.StarExpr); isStarExpr {
		//			return "*" + nameFromExpr(starExpr.X)
		//		} else if indexExpr, isIndexExpr := expr.(*ast.IndexExpr); isIndexExpr {
		//			return nameFromExpr(indexExpr.X) + "[" + nameFromExpr(indexExpr.Index) + "]"
		//		} else if indexListExpr, isIndexListExpr := expr.(*ast.IndexListExpr); isIndexListExpr {
		//			indices := []string{}
		//			for _, index := range indexListExpr.Indices {
		//				indices = append(indices, nameFromExpr(index))
		//			}
		//			return nameFromExpr(indexListExpr.X) + "[" + strings.Join(indices, ", ") + "]"
		//		} else if _, isArrayType := expr.(*ast.ArrayType); isArrayType {
		//			// something like `([]int)(nil)` or `[]byte("test")`
		//			return ""
		//		//} else if _, isBinaryExpr := expr.(*ast.BinaryExpr); isBinaryExpr {
		//		//	return ""
		//		//} else if _, isBasicLit := expr.(*ast.BasicLit); isBasicLit {
		//		//	return ""
		//		} else {
		//			panic(expr)
		//		}
		//	}
		//	name := nameFromExpr(callExpr.Fun)
		//	if name != "" {
		//		if _, not := calls[name]; !not {
		//			calls[name] = 0
		//		}
		//		calls[name] += 1
		//	}
		//	return true
		//})
		//Dump2(calls)
	}

//   src := `
//package main
//import "github.com/google/go-cmp/cmp"
//func a(int) string { return "test" }
//func main() {
//   b := a(8)
//   c := []int{1}
//   d := c[0]
//	if false { panic([]any{b, d}) }
//	if false {
//		e := "test_e"
//		d := "test_d"
//		if false { panic([]any{e, d}) }
//	}
//}
//`
//   fset := token.NewFileSet()
//   node := MustFirst(parser.ParseFile(fset, "", src, parser.AllErrors))
//   config := types.Config{Importer: nil} // use the default importer
//   info := &types.Info{
//		Types: make(map[ast.Expr]types.TypeAndValue),
//		//Scopes: make(map[ast.Node]*types.Scope),
//	}
//   MustFirst(config.Check("example", fset, []*ast.File{node}, info))
//   for expr, tv := range info.Types {
//       //fmt.Printf("Expression: %v, Type: %v\n", expr, tv.Type.String())
//       fmt.Printf("Expression: %v, Type: %v\n", src[expr.Pos()-1:expr.End()-1], tv.Type.String())
//   }
//	Dump2(info)
//	//Dump3(info)
}
