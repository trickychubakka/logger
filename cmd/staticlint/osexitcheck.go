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
