// Package order provides support for describing the ordering of data.
package order

import (
	"errors"
	"fmt"
	"strings"
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

// =============================================================================

// By represents a field used to order by and direction.
type By struct {
	Field     string
	Direction string
}

// IsZeroValue checks if the By value is empty.
func (b By) IsZeroValue() bool {
	return b == By{}
}

// =============================================================================

// Order represents a set of fields that represent the allowable fields for
// ordering and a default order field.
type Order struct {
	Fields  map[string]bool
	Default string
}

// New constructs a new Order with a specified set of fields.
func New(fields map[string]bool, defaultField string) *Order {
	return &Order{
		Fields:  fields,
		Default: defaultField,
	}
}

// Check validates the order by contains expected values.
func (o *Order) Check(by By) error {
	if _, exists := o.Fields[by.Field]; !exists {
		return fmt.Errorf("field %q is not a field you can order by", by.Field)
	}

	if _, exists := directions[by.Direction]; !exists {
		return fmt.Errorf("direction %q is not a value you can set order by", by.Direction)
	}

	return nil
}

// By constructs a By value from the specified parameters.
func (o *Order) By(field string, direction string) (By, error) {
	if _, exists := o.Fields[field]; !exists {
		return By{}, fmt.Errorf("invalid field %q provided", field)
	}

	if _, exists := directions[direction]; !exists {
		return By{}, fmt.Errorf("invalid direction %q provided", direction)
	}

	return By{Field: field, Direction: direction}, nil
}

// OrderByFromQueryString constructs a By value by parsing a string in the
// form of [field,direction]. Normally a string from a query string.
func (o *Order) OrderByFromQueryString(s string) (By, error) {
	if s == "" {
		return By{Field: o.Default, Direction: ASC}, nil
	}

	orderParts := strings.Split(s, ",")

	var by By
	switch len(orderParts) {
	case 1:
		by = By{Field: strings.Trim(orderParts[0], " "), Direction: ASC}
	case 2:
		by = By{Field: strings.Trim(orderParts[0], " "), Direction: strings.Trim(orderParts[1], " ")}
	default:
		return By{}, errors.New("invalid ordering information")
	}

	return by, nil
}
