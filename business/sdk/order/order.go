// Package order provides support for describing the ordering of data.
package order

import (
	"errors"
	"fmt"
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

// Parse constructs a By value by parsing a string in the form of
// "field,direction" ie "user_id,ASC".
func Parse(fieldMappings map[string]string, orderBy string, defaultOrder By) (By, error) {
	if orderBy == "" {
		return defaultOrder, nil
	}

	orderParts := strings.Split(orderBy, ",")

	orgFieldName := strings.TrimSpace(orderParts[0])
	fieldName, exists := fieldMappings[orgFieldName]
	if !exists {
		return By{}, validate.NewFieldsError(orgFieldName, errors.New("order field does not exist"))
	}

	switch len(orderParts) {
	case 1:
		return NewBy(fieldName, ASC), nil

	case 2:
		direction := strings.TrimSpace(orderParts[1])
		if _, exists := directions[direction]; !exists {
			return By{}, validate.NewFieldsError(orderBy, fmt.Errorf("unknown direction: %s", direction))
		}

		return NewBy(fieldName, direction), nil

	default:
		return By{}, validate.NewFieldsError(orderBy, errors.New("unknown order field"))
	}
}
