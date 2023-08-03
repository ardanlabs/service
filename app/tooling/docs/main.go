package main

import (
	"encoding/json"
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
		fmt.Print("\n============================\n\n")

		fmt.Printf("Route  : (%s) %s\n", war.method, war.route)
		fmt.Printf("Status : %s (%d)\n", war.status, statuses[war.status])

		for _, comment := range war.comments {
			fmt.Println(comment)
		}

		fmt.Print("\n")
		fmt.Println("Input Model :", produceJSONDocument(war.inputDoc))
		fmt.Print("\n")
		fmt.Println("Output Model :", produceJSONDocument(war.outputDoc))
		fmt.Print("\n")
		fmt.Printf("Paging Vars    : %v\n", strings.Join(war.queryVars.paging, ", "))
		fmt.Printf("Filtering Vars : %v\n", strings.Join(war.queryVars.filtering, ", "))
	}

	return nil
}

// =============================================================================

type field struct {
	Name     string
	Type     string
	Tag      string
	Optional bool
}

type queryVars struct {
	paging    []string
	filtering []string
}

type webAPIRecord struct {
	group     string
	tag       string
	method    string
	route     string
	status    string
	inputDoc  []field
	outputDoc []field
	comments  []string
	queryVars queryVars
}

func findWebAPIRecords() ([]webAPIRecord, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "app/services/sales-api/handlers/v1/productgrp/productgrp.go", nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("ParseFile: %w", err)
	}

	var wars []webAPIRecord

	f := func(n ast.Node) bool {

		// We only care if this node is a function declaration.
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		found, war, errF := parseWebAPI(fset, file, funcDecl, "productgrp")
		if found {
			wars = append(wars, war)
			return true
		}

		if errF != nil {
			err = fmt.Errorf("parseWebAPI: %w", err)
			return false
		}

		return true
	}

	ast.Inspect(file, f)

	return wars, err
}

func parseWebAPI(fset *token.FileSet, file *ast.File, funcDecl *ast.FuncDecl, group string) (found bool, war webAPIRecord, err error) {

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
		return false, webAPIRecord{}, nil
	}

	// Capture the last comment.
	comment := cGroup.List[len(cGroup.List)-1].Text

	// Does this comment have the webapi tag?
	const tag = "webapi"
	idx := strings.Index(comment, tag)
	if idx == -1 {
		return false, webAPIRecord{}, nil
	}

	// Split this comment by the space delimiter.
	record := strings.Split(comment[idx:], " ")

	// Capture any remaining comments that are not
	// part of the webapi tag.
	var comments []string
	for _, com := range cGroup.List[:len(cGroup.List)-1] {
		comments = append(comments, com.Text[3:])
	}

	// Create a webAPIRecord and assign what we have now.
	war = webAPIRecord{
		group:    group,
		tag:      strings.TrimSpace(record[0]),
		comments: comments,
	}

	// Match the key to the field in the webAPIRecord.
	for _, rec := range record {
		kv := strings.Split(rec, "=")

		switch kv[0] {
		case "method":
			war.method = kv[1]

		case "route":
			war.route = kv[1]

		case "inputdoc":
			inputDoc, err := findAppModel(group, kv[1])
			if err != nil {
				return false, war, fmt.Errorf("findAppModel input: %w", err)
			}
			war.inputDoc = inputDoc

		case "outputdoc":
			outputDoc, err := findAppModel(group, kv[1])
			if err != nil {
				return false, war, fmt.Errorf("findAppModel output: %w", err)
			}
			war.outputDoc = outputDoc

		case "status":
			war.status = kv[1]
		}
	}

	queryVars, err := findQueryVars(funcDecl.Body, group)
	if err != nil {
		return false, war, fmt.Errorf("findPageFilterOrder: %w", err)
	}
	war.queryVars = queryVars

	return true, war, nil
}

// =============================================================================

func findAppModel(group string, modelName string) ([]field, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "app/services/sales-api/handlers/v1/"+group+"/model.go", nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("ParseFile: %w", err)
	}

	var fields []field

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
			fields = append(fields, field{
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

func produceJSONDocument(fields []field) string {
	m := make(map[string]any)

	for _, field := range fields {
		tag := field.Tag
		typ := field.Type

		if field.Optional {
			tag = "*" + tag
		}

		if strings.Contains(strings.ToLower(field.Name), "id") {
			typ = "UUID"
		}

		if strings.Contains(strings.ToLower(field.Name), "date") {
			typ = "RFC3339"
		}

		m[tag] = typ
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return ""
	}

	doc := string(data)
	doc = strings.ReplaceAll(doc, "\"float64\"", "float64")
	doc = strings.ReplaceAll(doc, "\"int\"", "int")

	return doc
}

// =============================================================================

func findQueryVars(body *ast.BlockStmt, group string) (queryVars, error) {
	var qv queryVars

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
			qv.paging = append(qv.paging, "page", "rows")

		case "parseFilter":
			filtering, err := findFilters(qv.filtering, group)
			if err != nil {
				return queryVars{}, fmt.Errorf("findFilters: %w", err)
			}
			qv.filtering = filtering

		case "parseOrder":
		}
	}

	return qv, nil
}

func findFilters(queryVars []string, group string) ([]string, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "app/services/sales-api/handlers/v1/"+group+"/filter.go", nil, parser.ParseComments)
	if err != nil {
		return queryVars, fmt.Errorf("ParseFile: %w", err)
	}

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
						queryVars = append(queryVars, strings.Trim(bl.Value, "\""))
					}
				}
			}

			break
		}

		return true
	}

	ast.Inspect(file, f)

	return queryVars, nil
}

// =============================================================================

var statuses = map[string]int{
	"StatusContinue":                      100, // RFC 9110, 15.2.1
	"StatusSwitchingProtocols":            101, // RFC 9110, 15.2.2
	"StatusProcessing":                    102, // RFC 2518, 10.1
	"StatusEarlyHints":                    103, // RFC 8297
	"StatusOK":                            200, // RFC 9110, 15.3.1
	"StatusCreated":                       201, // RFC 9110, 15.3.2
	"StatusAccepted":                      202, // RFC 9110, 15.3.3
	"StatusNonAuthoritativeInfo":          203, // RFC 9110, 15.3.4
	"StatusNoContent":                     204, // RFC 9110, 15.3.5
	"StatusResetContent":                  205, // RFC 9110, 15.3.6
	"StatusPartialContent":                206, // RFC 9110, 15.3.7
	"StatusMultiStatus":                   207, // RFC 4918, 11.1
	"StatusAlreadyReported":               208, // RFC 5842, 7.1
	"StatusIMUsed":                        226, // RFC 3229, 10.4.1
	"StatusMultipleChoices":               300, // RFC 9110, 15.4.1
	"StatusMovedPermanently":              301, // RFC 9110, 15.4.2
	"StatusFound":                         302, // RFC 9110, 15.4.3
	"StatusSeeOther":                      303, // RFC 9110, 15.4.4
	"StatusNotModified":                   304, // RFC 9110, 15.4.5
	"StatusUseProxy":                      305, // RFC 9110, 15.4.6
	"StatusTemporaryRedirect":             307, // RFC 9110, 15.4.8
	"StatusPermanentRedirect":             308, // RFC 9110, 15.4.9
	"StatusBadRequest":                    400, // RFC 9110, 15.5.1
	"StatusUnauthorized":                  401, // RFC 9110, 15.5.2
	"StatusPaymentRequired":               402, // RFC 9110, 15.5.3
	"StatusForbidden":                     403, // RFC 9110, 15.5.4
	"StatusNotFound":                      404, // RFC 9110, 15.5.5
	"StatusMethodNotAllowed":              405, // RFC 9110, 15.5.6
	"StatusNotAcceptable":                 406, // RFC 9110, 15.5.7
	"StatusProxyAuthRequired":             407, // RFC 9110, 15.5.8
	"StatusRequestTimeout":                408, // RFC 9110, 15.5.9
	"StatusConflict":                      409, // RFC 9110, 15.5.10
	"StatusGone":                          410, // RFC 9110, 15.5.11
	"StatusLengthRequired":                411, // RFC 9110, 15.5.12
	"StatusPreconditionFailed":            412, // RFC 9110, 15.5.13
	"StatusRequestEntityTooLarge":         413, // RFC 9110, 15.5.14
	"StatusRequestURITooLong":             414, // RFC 9110, 15.5.15
	"StatusUnsupportedMediaType":          415, // RFC 9110, 15.5.16
	"StatusRequestedRangeNotSatisfiable":  416, // RFC 9110, 15.5.17
	"StatusExpectationFailed":             417, // RFC 9110, 15.5.18
	"StatusTeapot":                        418, // RFC 9110, 15.5.19 (Unused)
	"StatusMisdirectedRequest":            421, // RFC 9110, 15.5.20
	"StatusUnprocessableEntity":           422, // RFC 9110, 15.5.21
	"StatusLocked":                        423, // RFC 4918, 11.3
	"StatusFailedDependency":              424, // RFC 4918, 11.4
	"StatusTooEarly":                      425, // RFC 8470, 5.2.
	"StatusUpgradeRequired":               426, // RFC 9110, 15.5.22
	"StatusPreconditionRequired":          428, // RFC 6585, 3
	"StatusTooManyRequests":               429, // RFC 6585, 4
	"StatusRequestHeaderFieldsTooLarge":   431, // RFC 6585, 5
	"StatusUnavailableForLegalReasons":    451, // RFC 7725, 3
	"StatusInternalServerError":           500, // RFC 9110, 15.6.1
	"StatusNotImplemented":                501, // RFC 9110, 15.6.2
	"StatusBadGateway":                    502, // RFC 9110, 15.6.3
	"StatusServiceUnavailable":            503, // RFC 9110, 15.6.4
	"StatusGatewayTimeout":                504, // RFC 9110, 15.6.5
	"StatusHTTPVersionNotSupported":       505, // RFC 9110, 15.6.6
	"StatusVariantAlsoNegotiates":         506, // RFC 2295, 8.1
	"StatusInsufficientStorage":           507, // RFC 4918, 11.5
	"StatusLoopDetected":                  508, // RFC 5842, 7.2
	"StatusNotExtended":                   510, // RFC 2774, 7
	"StatusNetworkAuthenticationRequired": 511, // RFC 6585, 6
}
