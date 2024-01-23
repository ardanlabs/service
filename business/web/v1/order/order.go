// Package order provides support for describing the ordering of data.
package order

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ardanlabs/service/foundation/validate"
)

// Set of directions for data ordering.
const (
	ASC  = "ASC"
	DESC = "DESC"
)

var directions = map[string]string{
	ASC:  "ASC",
	DESC: "DESC",
}

// By represents a field used to order by and direction.
type By struct {
	Field     string
	Direction string
}

// NewBy constructs a new By value with no checks.
func NewBy(field string, direction string) By {
	if _, exists := directions[direction]; !exists {
		return By{
			Field:     field,
			Direction: ASC,
		}
	}

	return By{
		Field:     field,
		Direction: direction,
	}
}

// Parse constructs a By value by parsing a string in the form
// of "field,direction".
func Parse(r *http.Request, defaultOrder By) (By, error) {
	v := r.URL.Query().Get("orderBy")

	if v == "" {
		return defaultOrder, nil
	}

	orderParts := strings.Split(v, ",")

	var by By
	switch len(orderParts) {
	case 1:
		by = NewBy(strings.TrimSpace(orderParts[0]), ASC)

	case 2:
		direction := strings.Trim(orderParts[1], " ")
		if _, exists := directions[direction]; !exists {
			return By{}, validate.NewFieldsError(v, fmt.Errorf("unknown direction: %s", by.Direction))
		}

		by = NewBy(strings.Trim(orderParts[0], " "), direction)

	default:
		return By{}, validate.NewFieldsError(v, errors.New("unknown order field"))
	}

	return by, nil
}
