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
		fmt.Println("Route:", war.route)
		fmt.Println("Method:", war.method)
		for _, comment := range war.comments {
			fmt.Println(comment)
		}
		fmt.Print("\n")

		inputFields, err := findAppModel("productgrp", war.inputDoc)
		if err != nil {
			return fmt.Errorf("findAppModel, %w", err)
		}

		outputFields, err := findAppModel("productgrp", war.outputDoc)
		if err != nil {
			return fmt.Errorf("findAppModel, %w", err)
		}

		fmt.Println("Input Model:", war.inputDoc)
		fmt.Print(produceJSONDocument(inputFields), "\n\n")

		fmt.Println("Output Model", war.outputDoc)
		fmt.Print(produceJSONDocument(outputFields), "\n\n")

		fmt.Printf("Status %s(%d)\n", war.status, statuses[war.status])

		fmt.Print("\n============================\n\n")
	}

	return nil
}

// =============================================================================

type apiField struct {
	Name     string
	Type     string
	Tag      string
	Optional bool
}

func produceJSONDocument(fields []apiField) string {
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

func findAppModel(group string, modelName string) ([]apiField, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "app/services/sales-api/handlers/v1/"+group+"/model.go", nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("ParseFile: %w", err)
	}

	var fields []apiField

	f := func(n ast.Node) bool {

		// We only care to look at types.
		typeSpec, ok := n.(*ast.TypeSpec)
		if ok {

			// Did we find the model that was specified in the call?
			if typeSpec.Name.Name == modelName {
				structType := typeSpec.Type.(*ast.StructType)

				// Walk through the list of fields in this struct.
				for _, field := range structType.Fields.List {
					var fieldType *ast.Ident
					var optional bool

					// This is complicated. A field can be using pointer or
					// value semantics. There is a different type depending.
					// So we start by asking if the field is using pointer
					// semantics.
					starType, ok := field.Type.(*ast.StarExpr)

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
						fieldType, ok = field.Type.(*ast.Ident)
					}

					// We need to check that either type assersion succeeed.
					// Now look for the json tag on the field to know what
					// actual field name is being used after marshaling.
					if ok {

						// We will use the field name by default.
						tag := field.Names[0].Name

						// Check if there is a json tag and if so, parse
						// out the field name.
						idx := strings.Index(field.Tag.Value, "json")
						if idx != -1 {
							idx2 := strings.Index(field.Tag.Value[idx:], "\"")
							idx3 := idx + idx2 + 1
							idx4 := strings.Index(field.Tag.Value[idx3:], "\"")
							tag = field.Tag.Value[idx3 : idx3+idx4]
						}

						// Add the field information to the list.
						fields = append(fields, apiField{
							Name:     field.Names[0].Name,
							Type:     fieldType.Name,
							Tag:      tag,
							Optional: optional,
						})
					}
				}
			}
		}

		return true
	}

	ast.Inspect(file, f)

	return fields, nil
}

// =============================================================================

type webAPIRecord struct {
	tag       string
	method    string
	route     string
	inputDoc  string
	outputDoc string
	status    string
	comments  []string
}

func findWebAPIRecords() ([]webAPIRecord, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "app/services/sales-api/handlers/v1/productgrp/productgrp.go", nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("ParseFile: %w", err)
	}

	var tag = "webapi"
	var wars []webAPIRecord

	f := func(n ast.Node) bool {

		// We only care to look at functions.
		funcDecl, ok := n.(*ast.FuncDecl)
		if ok {

			// Capture the line number for this function declaration.
			line := fset.Position(funcDecl.Pos()).Line

			// Search through the comments for this function.
			for _, cGroup := range file.Comments {

				// We need the last comment to check for the webapi tag.
				if fset.Position(cGroup.End()).Line == line-1 {

					// Capture the last comment.
					comment := cGroup.List[len(cGroup.List)-1].Text

					// Does this comment have the webapi tag?
					if n := strings.Index(comment, tag); n != -1 {

						// Split this comment by the space delimiter.
						record := strings.Split(comment[n:], " ")

						// Capture any remaining comments that are not
						// part of the webapi tag.
						var comments []string
						for _, com := range cGroup.List[:len(cGroup.List)-1] {
							comments = append(comments, com.Text[3:])
						}

						// Create a webAPIRecord and assign what we have now.
						war := webAPIRecord{
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
								war.inputDoc = kv[1]
							case "outputdoc":
								war.outputDoc = kv[1]
							case "status":
								war.status = kv[1]
							}
						}

						// Append this webAPIRecord to the list.
						wars = append(wars, war)
					}
				}
			}
		}

		return true
	}

	ast.Inspect(file, f)

	return wars, nil
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
