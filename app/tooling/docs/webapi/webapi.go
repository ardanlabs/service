// Package webapi provides support for extracting web api information from
// reading the code.
package webapi

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// Records produces and returns the web api information for the specified
// handler group.
func Records(group string) ([]Record, error) {
	fset := token.NewFileSet()

	fileName := fmt.Sprintf("app/services/sales-api/handlers/v1/%s/%s.go", group, group)
	file, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("ParseFile: %w", err)
	}

	var records []Record

	f := func(n ast.Node) bool {

		// We only care if this node is a function declaration.
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		found, record, errF := parseWebAPI(fset, file, funcDecl, group)
		if found {
			records = append(records, record)
			return true
		}

		if errF != nil {
			err = fmt.Errorf("parseWebAPI: %w", err)
			return false
		}

		return true
	}

	ast.Inspect(file, f)

	return records, err
}

// =============================================================================

func parseWebAPI(fset *token.FileSet, file *ast.File, funcDecl *ast.FuncDecl, group string) (bool, Record, error) {

	// Capture the line number for this function declaration.
	line := fset.Position(funcDecl.Pos()).Line

	// Search through the group of comments in this file looking for a
	// comment that exist in the line above the function declaration.
	var cGroup *ast.CommentGroup
	for _, cGroup = range file.Comments {

		// We are looking for the comments associated with this function.
		if fset.Position(cGroup.End()).Line == line-1 {
			break
		}
	}

	// We didn't find any comments.
	if cGroup == nil {
		return false, Record{}, nil
	}

	// Capture the last comment.
	comment := cGroup.List[len(cGroup.List)-1].Text

	// Does this comment have the webapi tag?
	const tag = "webapi"
	idx := strings.Index(comment, tag)
	if idx == -1 {
		return false, Record{}, nil
	}

	// Capture any remaining comments that are not
	// part of the webapi tag.
	var comments []string
	for _, com := range cGroup.List[:len(cGroup.List)-1] {
		comments = append(comments, com.Text[3:])
	}

	// Create a webAPIRecord and assign what we have now.
	record := Record{
		Group:    group,
		Comments: comments,
	}

	// Split this comment by the space delimiter.
	fields := strings.Split(comment[idx:], " ")

	// Match the key to the field in the webAPIRecord.
	for _, field := range fields {
		kv := strings.Split(field, "=")

		switch kv[0] {
		case "method":
			record.Method = kv[1]

		case "route":
			record.Route = kv[1]

		case "inputdoc":
			inputDoc, err := findAppModel(group, kv[1])
			if err != nil {
				return false, Record{}, fmt.Errorf("findAppModel input: %w", err)
			}
			record.InputDoc = inputDoc

		case "outputdoc":
			outputDoc, err := findAppModel(group, kv[1])
			if err != nil {
				return false, Record{}, fmt.Errorf("findAppModel output: %w", err)
			}
			record.OutputDoc = outputDoc

		case "status":
			record.Status = kv[1]
		}
	}

	queryVars, err := findQueryVars(funcDecl.Body, group)
	if err != nil {
		return false, Record{}, fmt.Errorf("findPageFilterOrder: %w", err)
	}
	record.QueryVars = queryVars

	return true, record, nil
}

// =============================================================================

func findAppModel(group string, modelName string) ([]Field, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "app/services/sales-api/handlers/v1/"+group+"/model.go", nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("ParseFile: %w", err)
	}

	var fields []Field

	f := func(n ast.Node) bool {

		// We only care to look at types.
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		// Did we find the model that was specified in the call?
		if typeSpec.Name.Name != modelName {
			return true
		}

		structType := typeSpec.Type.(*ast.StructType)

		// Walk through the list of fields in this struct.
		for _, astField := range structType.Fields.List {
			var fieldType *ast.Ident
			var optional bool

			// This is complicated. A field can be using pointer or
			// value semantics. There is a different type depending.
			// So we start by asking if the field is using pointer
			// semantics.
			starType, ok := astField.Type.(*ast.StarExpr)

			// If this field was using pointer semantics, then we
			// use the starType variable to get to the identifier
			// and mark this field as optional.
			//
			// If this field was using value semantics, then we
			// use the field variable to get to the identifier.
			switch ok {
			case true:
				fieldType, ok = starType.X.(*ast.Ident)
				optional = true
			default:
				fieldType, ok = astField.Type.(*ast.Ident)
			}

			// We need to check that either type assersion succeeed.
			if !ok {
				continue
			}

			// Now look for the json tag on the field to know what
			// actual field name is being used after marshaling.

			// We will use the field name by default.
			tag := astField.Names[0].Name

			// Check if there is a json tag and if so, parse
			// out the field name.
			idx := strings.Index(astField.Tag.Value, "json")
			if idx != -1 {
				idx2 := strings.Index(astField.Tag.Value[idx:], "\"")
				idx3 := idx + idx2 + 1
				idx4 := strings.Index(astField.Tag.Value[idx3:], "\"")
				tag = astField.Tag.Value[idx3 : idx3+idx4]
			}

			// Add the field information to the list.
			fields = append(fields, Field{
				Name:     astField.Names[0].Name,
				Type:     fieldType.Name,
				Tag:      tag,
				Optional: optional,
			})
		}

		return true
	}

	ast.Inspect(file, f)

	return fields, nil
}

// =============================================================================

func findQueryVars(body *ast.BlockStmt, group string) (QueryVars, error) {
	var qv QueryVars

	// Walk through the body of the function looking for calls to
	// paging, parseFilter, and parseOrder.
	for _, stmt := range body.List {

		// Start by looking for assignment statements.
		agn, ok := stmt.(*ast.AssignStmt)
		if !ok {
			continue
		}

		// If a function call is not being made, ignore.
		ce, ok := agn.Rhs[0].(*ast.CallExpr)
		if !ok {
			continue
		}

		var ident *ast.Ident

		// We might have a method call (*ast.SelectorExpr) or
		// function call (*ast.Ident). Check if we have a method
		// call first.
		se, ok := ce.Fun.(*ast.SelectorExpr)

		// If we had a method call, then use the X field to get
		// to the identifier information. If this was a function
		// call, then use the Fun field from the call expression.
		switch ok {
		case true:
			ident, ok = se.X.(*ast.Ident)
		default:
			ident, ok = ce.Fun.(*ast.Ident)
		}

		// We need to check that either type assersion succeeed.
		if !ok {
			continue
		}

		switch ident.Name {
		case "paging":
			qv.Paging = append(qv.Paging, "page", "rows")

		case "parseFilter":
			filterBy, err := findFilters(group)
			if err != nil {
				return QueryVars{}, fmt.Errorf("findFilters: %w", err)
			}
			qv.FilterBy = filterBy

		case "parseOrder":
			orderBy, err := findOrders(group)
			if err != nil {
				return QueryVars{}, fmt.Errorf("findOrders: %w", err)
			}
			qv.OrderBy = orderBy
		}
	}

	return qv, nil
}

func findFilters(group string) ([]string, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "app/services/sales-api/handlers/v1/"+group+"/filter.go", nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("ParseFile: %w", err)
	}

	var filterBy []string

	f := func(n ast.Node) bool {

		// We only care if this node is a function declaration.
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		// Is this function the parseFilter function.
		if funcDecl.Name.Name != "parseFilter" {
			return true
		}

		// We need to find all the value.Get calls.
		for _, stmt := range funcDecl.Body.List {

			// We only care if this node is a value spec.
			vs, ok := stmt.(*ast.DeclStmt)
			if !ok {
				continue
			}

			gd, ok := vs.Decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			for _, sp := range gd.Specs {
				vs, ok := sp.(*ast.ValueSpec)
				if !ok {
					break
				}

				for i, n := range vs.Names {
					if !strings.Contains(n.Name, "filterBy") {
						continue
					}

					// Capture the value assigned to the constant.
					bl, ok := vs.Values[i].(*ast.BasicLit)
					if ok {
						filterBy = append(filterBy, strings.Trim(bl.Value, "\""))
					}
				}
			}

			break
		}

		return true
	}

	ast.Inspect(file, f)

	return filterBy, nil
}

func findOrders(group string) ([]string, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "app/services/sales-api/handlers/v1/"+group+"/order.go", nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("ParseFile: %w", err)
	}

	var orderBy []string

	f := func(n ast.Node) bool {

		// We only care if this node is a function declaration.
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		// Is this function the parseOrder function.
		if funcDecl.Name.Name != "parseOrder" {
			return true
		}

		// We need to find all the value.Get calls.
		for _, stmt := range funcDecl.Body.List {

			// We only care if this node is a value spec.
			vs, ok := stmt.(*ast.DeclStmt)
			if !ok {
				continue
			}

			gd, ok := vs.Decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			for _, sp := range gd.Specs {
				vs, ok := sp.(*ast.ValueSpec)
				if !ok {
					break
				}

				for i, n := range vs.Names {
					if !strings.Contains(n.Name, "orderBy") {
						continue
					}

					// Capture the value assigned to the constant.
					bl, ok := vs.Values[i].(*ast.BasicLit)
					if ok {
						orderBy = append(orderBy, strings.Trim(bl.Value, "\""))
					}
				}
			}

			break
		}

		return true
	}

	ast.Inspect(file, f)

	return orderBy, nil
}
