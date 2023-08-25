package webapi

import (
	"strings"
)

// toModel covnerts a collection of fields to a model document.
func toModel(fields []Field, slice bool) any {
	if len(fields) == 0 {
		return nil
	}

	m := make(map[string]any)

	for _, field := range fields {
		tag := field.Tag
		typ := field.Type

		if tag == "-" {
			continue
		}

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

	if slice {
		return []map[string]any{
			m,
		}
	}

	return m
}
