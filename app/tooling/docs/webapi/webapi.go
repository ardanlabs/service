// Package webapi provides support for extracting web api information from
// reading the code.
package webapi

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"strings"
)

var methods = map[string]string{
	"MethodDelete":  http.MethodDelete,
	"MethodGet":     http.MethodGet,
	"MethodHead":    http.MethodHead,
	"MethodOptions": http.MethodOptions,
	"MethodPatch":   http.MethodPatch,
	"MethodPost":    http.MethodPost,
	"MethodPut":     http.MethodPut,
	"MethodTrace":   http.MethodTrace,
}

func Routes(version string) ([]Route, error) {
	dirEntries, err := os.ReadDir("app/services/sales-api/v1/handlers")
	if err != nil {
		return nil, fmt.Errorf("ReadDir: %w", err)
	}

	var files []struct {
		group string
		file  string
	}

	for _, entry := range dirEntries {
		if !entry.IsDir() {
			continue
		}

		files = append(files, struct {
			group string
			file  string
		}{
			entry.Name(),
			fmt.Sprintf("app/services/sales-api/v1/handlers/%s/route.go", entry.Name()),
		})
	}

	var routes []Route

	for _, item := range files {
		fset := token.NewFileSet()

		file, err := parser.ParseFile(fset, item.file, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("ParseFile: %w", err)
		}

		f := func(n ast.Node) bool {

			// We only care if this node is a function declaration.
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			// We only want the routes function.
			if funcDecl.Name.Name != "Routes" {
				return false
			}

			var ver string

			// We need to find all the value.Get calls.
			for _, stmt := range funcDecl.Body.List {

				// We are looking for the version of this route.
				if v, ok := stmt.(*ast.DeclStmt); ok {
					ver = v.Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Values[0].(*ast.BasicLit).Value
					continue
				}

				// We are looking for expressions that will represent the
				// call to Handle.
				es, ok := stmt.(*ast.ExprStmt)
				if !ok {
					continue
				}

				ce, ok := es.X.(*ast.CallExpr)
				if !ok {
					continue
				}

				se, ok := ce.Fun.(*ast.SelectorExpr)
				if !ok {
					continue
				}

				if se.Sel.Name != "Handle" {
					continue
				}

				// The first parameter to Handle represents the method information.
				method, ok := ce.Args[0].(*ast.SelectorExpr)
				if !ok {
					continue
				}

				// The third parameter to Handle represents the route.
				url, ok := ce.Args[2].(*ast.BasicLit)
				if !ok {
					continue
				}

				// The forth parameter to Handle represents the name of the
				// handler function.
				handler, ok := ce.Args[3].(*ast.SelectorExpr)
				if !ok {
					continue
				}

				routes = append(routes, Route{
					Method:   methods[method.Sel.Name],
					URL:      "/" + ver[:len(ver)-1] + url.Value[1:],
					Handler:  handler.Sel.Name,
					Group:    item.group,
					ErrorDoc: "ErrorResponse",
					File:     fmt.Sprintf("app/services/sales-api/%s/handlers/%s/%s.go", version, item.group, item.group),
				})
			}

			return true
		}

		ast.Inspect(file, f)
	}

	return routes, nil
}

// Records produces and returns the web api information for the specified
// handler group.
func Records(routes []Route) ([]Record, error) {
	var records []Record

	for _, route := range routes {
		fset := token.NewFileSet()

		file, err := parser.ParseFile(fset, route.File, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("ParseFile: %w", err)
		}

		f := func(n ast.Node) bool {

			// We only care if this node is a function declaration.
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			// We are looking for this route.
			if funcDecl.Name.Name != route.Handler {
				return true
			}

			found, record, errF := parseWebAPI(fset, file, funcDecl, route)
			if found {
				records = append(records, record)
				return true
			}

			if errF != nil {
				err = fmt.Errorf("parseWebAPI: %w", errF)
				return false
			}

			return true
		}

		ast.Inspect(file, f)

		if err != nil {
			return nil, fmt.Errorf("inspect: %w", err)
		}
	}

	return records, nil
}

func parseWebAPI(fset *token.FileSet, file *ast.File, funcDecl *ast.FuncDecl, route Route) (bool, Record, error) {

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

	// Capture any remaining comments that are not
	// part of the webapi tag.
	var comments []string
	for _, com := range cGroup.List {
		comments = append(comments, com.Text[3:])
	}

	// Create a webAPIRecord and assign what we have now.
	record := Record{
		Group:    route.Group,
		Method:   route.Method,
		Route:    route.URL,
		Comments: comments,
	}

	// Find the error document.
	errorDoc, err := findAppModel(route.Group, route.ErrorDoc)
	if err != nil {
		return false, Record{}, fmt.Errorf("findAppModel error: %w", err)
	}
	record.ErrorDoc = toModel(errorDoc, false)

	// Fing the input document.
	modelName := findInputDocument(funcDecl.Body)
	if modelName != "" {
		inputDoc, err := findAppModel(route.Group, modelName)
		if err != nil {
			return false, Record{}, fmt.Errorf("findAppModel input: %w", err)
		}
		record.InputDoc = toModel(inputDoc, false)
	}

	// Find the output document.
	funcName, status := findOutputDocument(funcDecl.Body)
	record.Status = status

	if funcName != "" {
		parts := strings.Split(funcName, ".")

		switch parts[0] {
		case "v1":
			funcName, _ = strings.CutPrefix(parts[1], "PageDocument[")
			funcName, _ = strings.CutSuffix(funcName, "]")

			concreteModelName, _, err := findAppFunctionFromModel(route.Group, funcName)
			if err != nil {
				return false, Record{}, fmt.Errorf("findAppModel output: %w", err)
			}

			concreteModel, err := findAppModel(route.Group, concreteModelName)
			if err != nil {
				return false, Record{}, fmt.Errorf("findAppModel output: %w", err)
			}

			outputDoc, err := findAppModel(route.Group, "PageDocument")
			if err != nil {
				return false, Record{}, fmt.Errorf("findAppModel output: %w", err)
			}

			outputDoc[0].Type = toModel(concreteModel, true)

			record.OutputDoc = toModel(outputDoc, false)

		default:
			modelName, slice, err := findAppFunctionFromModel(route.Group, funcName)
			if err != nil {
				return false, Record{}, fmt.Errorf("findAppModel output: %w", err)
			}

			outputDoc, err := findAppModel(route.Group, modelName)
			if err != nil {
				return false, Record{}, fmt.Errorf("findAppModel output: %w", err)
			}

			record.OutputDoc = toModel(outputDoc, slice)
		}
	}

	// Find the query vars.
	queryVars, err := findQueryVars(funcDecl.Body, route.Group)
	if err != nil {
		return false, Record{}, fmt.Errorf("findPageFilterOrder: %w", err)
	}
	record.QueryVars = queryVars

	return true, record, nil
}

// This function looks for the web.Respond call at the end of each handler.
// Once found, it returns the name of the type that is used in the trusted.
func findOutputDocument(body *ast.BlockStmt) (funcName string, status string) {

	// Walk through the body of the function looking for the web.Decode
	// function call.
	for _, stmt := range body.List {

		// Start by looking for a return statement.
		rs, ok := stmt.(*ast.ReturnStmt)
		if !ok {
			continue
		}

		// Look for a function call inside the return.
		ce, ok := rs.Results[0].(*ast.CallExpr)
		if !ok {
			continue
		}

		se, ok := ce.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}

		// We are looking for the web.Respond call.
		if se.Sel.Name != "Respond" {
			continue
		}

		// Now we need to find the forth parameter which should
		// be the status.
		if stat, ok := ce.Args[3].(*ast.SelectorExpr); ok {
			status = stat.Sel.Name
		}

		// Now we need to find the third parameter which should
		// be a to function.
		ce, ok = ce.Args[2].(*ast.CallExpr)
		if !ok {
			continue
		}

		var isPageDocument bool

		// Did we find PageDocument and if so, we need the
		// to function inside that call.
		if se, ok = ce.Fun.(*ast.SelectorExpr); ok {
			if se.Sel.Name != "NewPageDocument" {
				continue
			}
			ce, ok = ce.Args[0].(*ast.CallExpr)
			if !ok {
				continue
			}
			isPageDocument = true
		}

		// Did we find a to function.
		idt, ok := ce.Fun.(*ast.Ident)
		if !ok {
			continue
		}

		// This is the actual document that is being sent back with
		// the idt.Name type as the generic type.
		if isPageDocument {
			return fmt.Sprintf("v1.PageDocument[%s]", idt.Name), status
		}

		return idt.Name, status
	}

	return "", status
}

func findInputDocument(body *ast.BlockStmt) string {

	// Walk through the body of the function looking for the web.Decode
	// function call.
	for _, stmt := range body.List {

		// Start by looking for an if statement.
		ifs, ok := stmt.(*ast.IfStmt)
		if !ok {
			continue
		}

		// Look at the assignment inside the if statement.
		agn, ok := ifs.Init.(*ast.AssignStmt)
		if !ok {
			continue
		}

		// If a function call is not being made, ignore.
		ce, ok := agn.Rhs[0].(*ast.CallExpr)
		if !ok {
			continue
		}

		// We need to extract the name of the function being called.
		se, ok := ce.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}

		// If this is not the web.Decode call, move on.
		if se.Sel.Name != "Decode" {
			continue
		}

		// Now we need the name of the varaible being passed
		// to the decode call.
		ue, ok := ce.Args[1].(*ast.UnaryExpr)
		if !ok {
			continue
		}

		// We now have the name of the variable.
		idtVarName, ok := ue.X.(*ast.Ident)
		if !ok {
			continue
		}

		// Look again at this function for the declaration of this
		// variable being passed into web.Decode.
		for _, stmt2 := range body.List {
			ds, ok := stmt2.(*ast.DeclStmt)
			if !ok {
				continue
			}

			gd, ok := ds.Decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			vs, ok := gd.Specs[0].(*ast.ValueSpec)
			if !ok {
				continue
			}

			idt, ok := vs.Type.(*ast.Ident)
			if !ok {
				continue
			}

			// Did we find the declaration and the type information?
			if idtVarName.Name == vs.Names[0].Name {
				return idt.Name
			}
		}
	}

	return ""
}

func findAppFunctionFromModel(group string, funcName string) (string, bool, error) {
	fset := token.NewFileSet()

	var idt *ast.Ident
	var slice bool

	file, err := parser.ParseFile(fset, "app/services/sales-api/v1/handlers/"+group+"/model.go", nil, parser.ParseComments)
	if err != nil {
		return "", false, fmt.Errorf("ParseFile: %w", err)
	}

	f := func(n ast.Node) bool {
		if idt != nil {
			return false
		}

		// We only care to look at functions.
		fs, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		// Did we find the function we are looking for.
		if fs.Name.Name != funcName {
			return true
		}

		// If the return type is a struct type, set the ident.
		idt, ok = fs.Type.Results.List[0].Type.(*ast.Ident)
		if ok {
			return false
		}

		// If the return type is an array, we need one more step to
		// get the ident.
		at, ok := fs.Type.Results.List[0].Type.(*ast.ArrayType)
		if !ok {
			return true
		}

		idt, _ = at.Elt.(*ast.Ident)
		slice = true

		return false
	}

	ast.Inspect(file, f)

	if idt == nil {
		return "", false, nil
	}

	return idt.Name, slice, nil
}

func findAppModel(group string, modelName string) ([]Field, error) {
	fset := token.NewFileSet()

	var file *ast.File
	var err error

	switch {
	case strings.Contains(modelName, "ErrorResponse"):
		file, err = parser.ParseFile(fset, "business/web/v1/v1.go", nil, parser.ParseComments)
	case strings.Contains(modelName, "PageDocument"):
		file, err = parser.ParseFile(fset, "business/web/v1/v1.go", nil, parser.ParseComments)
	default:
		file, err = parser.ParseFile(fset, "app/services/sales-api/v1/handlers/"+group+"/model.go", nil, parser.ParseComments)
	}

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
			var fieldType string
			var optional bool

			// This is complicated. A field can be using pointer or
			// value semantics. There is a different type depending.
			// So we start by asking if the field is using pointer
			// semantics.
			starType, okStar := astField.Type.(*ast.StarExpr)
			arryType, okArray := astField.Type.(*ast.ArrayType)

			switch {
			case okStar:
				// If this field was using pointer semantics, so we
				// use the starType variable to get to the identifier
				// and mark this field as optional.
				v, ok := starType.X.(*ast.Ident)
				if !ok {
					continue
				}
				fieldType = v.Name
				optional = true

			case okArray:
				v, ok := arryType.Elt.(*ast.Ident)
				if !ok {
					continue
				}
				if !ok {
					continue
				}
				fieldType = v.Name

			default:
				// If this field was using value semantics, so we
				// use the field variable to get to the identifier.
				switch v := astField.Type.(type) {
				case *ast.Ident:
					fieldType = v.Name
				case *ast.MapType:
					keyType, ok := v.Key.(*ast.Ident)
					if !ok {
						continue
					}
					keyVal, ok := v.Value.(*ast.Ident)
					if !ok {
						continue
					}
					fieldType = fmt.Sprintf("map[%s]%s", keyType, keyVal)
				default:
					continue
				}
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
				Type:     fieldType,
				Tag:      tag,
				Optional: optional,
			})
		}

		return true
	}

	ast.Inspect(file, f)

	return fields, nil
}

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
		case "page":
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

	file, err := parser.ParseFile(fset, "app/services/sales-api/v1/handlers/"+group+"/filter.go", nil, parser.ParseComments)
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

	file, err := parser.ParseFile(fset, "app/services/sales-api/v1/handlers/"+group+"/order.go", nil, parser.ParseComments)
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
