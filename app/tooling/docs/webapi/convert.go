package webapi

import (
	"encoding/json"
	"strings"
)

// ToJSON covnerts a collection of fields to a JSON document.
func ToJSON(fields []Field) string {
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
