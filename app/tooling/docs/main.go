package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strings"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "app/services/sales-api/handlers/v1/productgrp/productgrp.go", nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("ParseFile: %w", err)
	}

	ast.Inspect(file, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if ok {
			line := fset.Position(funcDecl.Pos()).Line
			for _, cGroup := range file.Comments {
				if fset.Position(cGroup.End()).Line == line-1 {
					comment := cGroup.List[len(cGroup.List)-1].Text
					if strings.Contains(comment, "service:webapi") {
						fmt.Println(funcDecl.Name, comment[2:])
					}
				}
			}
		}
		return true
	})

	return nil
}
