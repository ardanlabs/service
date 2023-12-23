// Package text converts the webapi records into text output.
package text

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/ardanlabs/service/app/tooling/docs/webapi"
)

//go:embed template.txt
var document string

func Transform(records []webapi.Record) error {
	var funcMap = template.FuncMap{
		"minus":  minus,
		"status": status,
		"json":   toJSON,
	}

	tmpl := template.Must(template.New("webapi").Funcs(funcMap).Parse(document))

	var b bytes.Buffer
	err := tmpl.Execute(&b, records)
	if err != nil {
		return err
	}

	fmt.Print(b.String())

	return nil
}

func minus(a, b int) int {
	return a - b
}

func status(status string) int {
	return webapi.Statuses[status]
}

func toJSON(v any) string {
	data, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return ""
	}

	return string(data)
}
