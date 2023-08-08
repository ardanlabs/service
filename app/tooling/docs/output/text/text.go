package text

import (
	"bytes"
	_ "embed"
	"text/template"

	"github.com/ardanlabs/service/app/tooling/docs/webapi"
)

//go:embed template.txt
var document string

func Transform(records []webapi.Record) (string, error) {
	var funcMap = template.FuncMap{
		"minus":  minus,
		"status": status,
		"json":   webapi.ToJSON,
	}

	tmpl := template.Must(template.New("webapi").Funcs(funcMap).Parse(document))

	var b bytes.Buffer
	err := tmpl.Execute(&b, records)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

// =============================================================================

func minus(a, b int) int {
	return a - b
}

func status(status string) int {
	return webapi.Statuses[status]
}
