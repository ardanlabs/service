// Package json converts the webapi records into json.
package json

import (
	"encoding/json"

	"github.com/ardanlabs/service/app/tooling/docs/webapi"
)

// Transform converts the collection of webapi records into json.
func Transform(records []webapi.Record) (string, error) {
	data, err := json.MarshalIndent(records, "", "    ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}
