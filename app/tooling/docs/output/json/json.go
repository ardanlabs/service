// Package json converts the webapi records into json.
package json

import (
	"fmt"

	"github.com/ardanlabs/service/app/tooling/docs/webapi"
	"github.com/go-json-experiment/json"
)

// Transform converts the collection of webapi records into json.
func Transform(records []webapi.Record) error {
	data, err := json.Marshal(records)
	if err != nil {
		return err
	}

	fmt.Print(string(data))

	return nil
}
