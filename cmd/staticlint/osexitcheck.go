// Package staticlint -- multichecker, статический анализатор.
// Файл osexitcheck.go -- анализатор для поиска прямого вызова os.Exit.
package main

import (
	"bytes"
	"flag"
	"go/ast"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/analysis"
	"log"
)

var OSExitCheckAnalyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "check for os.Exit call",
	Run:  runOSExitSearch,
}

// printFuncName -- вывод названия функции в AST узле.
func printFuncName(fset *token.FileSet, x *ast.CallExpr) (string, error) {
	var buf bytes.Buffer
	err := printer.Fprint(&buf, fset, x)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// runOSExitSearch -- run функция, отвечающая за анализ исходного кода.
func runOSExitSearch(pass *analysis.Pass) (interface{}, error) {

	flag.Parse()
	fset := token.NewFileSet()

	for _, file := range pass.Files {
		//log.Println("AST file is :", file.Name)
		ast.Inspect(file, func(n ast.Node) bool {
			// проверяем, что лежащий в узле тип -- это вызов функции
			if v, ok := n.(*ast.CallExpr); ok {
				if s, err := printFuncName(fset, v); s == "os.Exit(1)" {
					pass.Reportf(v.Pos(), "called func %s", v.Fun)
				} else if err != nil {
					log.Println("runOSExit error:", err)
				}
			}
			return true
		})
	}
	//}
	return nil, nil
}

// todochecker
//var TODOCheckAnalyzer = &analysis.Analyzer{
//	Name: "todocheck",
//	Doc:  "check for todo comments",
//	Run:  runTODOSearch,
//}
//
//func runTODOSearch(pass *analysis.Pass) (interface{}, error) {
//	flag.Parse()
//	fset := token.NewFileSet()
//	log.Println(fset)
//	// TODO todo comment 1
//	log.Println("RUN TODO CHECKER")
//	for _, file := range pass.Files {
//		//log.Println("AST file is :", file.Name.Name)
//		ast.Inspect(file, func(n ast.Node) bool {
//			// проверяем, что лежащий в узле тип -- это комментарий.
//			if v, ok := n.(*ast.Comment); ok {
//				log.Println("runTODOSearch: found COMMENT", v.Text)
//				if strings.HasPrefix(v.Text, "// TODO") {
//					//log.Println("runTODOSearch: found TODO")
//					pass.Reportf(v.Pos(), "TODO comment Reportf: %s", v.Text)
//				}
//			}
//			return true
//		})
//	}
//	//}
//	return nil, nil
//}
//
//// filesTreeGoFiles -- функция формирования среза с названиями go-файлов в директории.
//func filesTreeGoFiles(d string) ([]string, error) {
//	var s []string
//	err := filepath.Walk(d,
//		func(path string, info os.FileInfo, err error) error {
//			if err != nil {
//				return err
//			}
//			//fmt.Println(path)
//			if !info.IsDir() && filepath.Ext(info.Name()) == ".go" {
//				s = append(s, path)
//			}
//			return nil
//		})
//	if err != nil {
//		log.Println(err)
//		return nil, err
//	}
//	return s, nil
//}
