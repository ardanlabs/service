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
	wars, err := findWebAPIRecords()
	if err != nil {
		return fmt.Errorf("findWebAPIRecords, %w", err)
	}

	for _, war := range wars {
		fmt.Printf("%#v\n", war)
	}

	return nil
}

// =============================================================================

type webAPIRecord struct {
	tag        string
	method     string
	route      string
	inputType  string
	outputType string
	status     string
}

func findWebAPIRecords() ([]webAPIRecord, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "app/services/sales-api/handlers/v1/productgrp/productgrp.go", nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("ParseFile: %w", err)
	}

	var tag = "service:webapi"
	var wars []webAPIRecord

	ast.Inspect(file, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if ok {
			line := fset.Position(funcDecl.Pos()).Line
			for _, cGroup := range file.Comments {
				if fset.Position(cGroup.End()).Line == line-1 {
					comment := cGroup.List[len(cGroup.List)-1].Text
					if n := strings.Index(comment, tag); n != -1 {
						record := strings.Split(comment[n:], " ")

						war := webAPIRecord{
							tag:        strings.TrimSpace(record[0]),
							method:     record[1],
							route:      record[2],
							inputType:  record[3],
							outputType: record[4],
							status:     record[5],
						}

						wars = append(wars, war)
					}
				}
			}
		}
		return true
	})

	return wars, nil
}
