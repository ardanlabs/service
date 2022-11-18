// Package sort provides support for describing the sorting and ordering of data.
package sort

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

// OrderBy represents a field to order by and direction.
type OrderBy struct {
	Field     string
	Direction string
}

// IsZeroValue checks if an order is set to its zero value.
func (ob OrderBy) IsZeroValue() bool {
	return ob == OrderBy{}
}

// =============================================================================

// Order represents a set of fields and the default field that is allowed for ordering.
type Order struct {
	Fields  map[string]bool
	Default string
}

// NewOrder constructs a new order with a specified set of fields.
func NewOrder(fields map[string]bool, defaultField string) *Order {
	return &Order{
		Fields:  fields,
		Default: defaultField,
	}
}

// Check validates the order by contains expected values.
func (o *Order) Check(orderBy OrderBy) error {
	if _, exists := o.Fields[orderBy.Field]; !exists {
		return fmt.Errorf("field %q is not a field you can order by", orderBy.Field)
	}

	if _, exists := directions[orderBy.Direction]; !exists {
		return fmt.Errorf("direction %q is not a value you can set order by", orderBy.Direction)
	}

	return nil
}

// NewOrderBy constructs an Order from the specified parameters.
func (o *Order) NewOrderBy(field string, direction string) (OrderBy, error) {
	if _, exists := o.Fields[field]; !exists {
		return OrderBy{}, fmt.Errorf("invalid field %q provided", field)
	}

	if _, exists := directions[direction]; !exists {
		return OrderBy{}, fmt.Errorf("invalid direction %q provided", direction)
	}

	return OrderBy{Field: field, Direction: direction}, nil
}

// FromQueryString takes a query string for ordering and creates a orderby
// value. Expected format is field or field,direction
func (o *Order) FromQueryString(orderQuery string) (OrderBy, error) {
	if orderQuery == "" {
		return OrderBy{Field: o.Default, Direction: ASC}, nil
	}

	orderParts := strings.Split(orderQuery, ",")

	var orderBy OrderBy
	switch len(orderParts) {
	case 1:
		orderBy = OrderBy{Field: strings.Trim(orderParts[0], " "), Direction: ASC}
	case 2:
		orderBy = OrderBy{Field: strings.Trim(orderParts[0], " "), Direction: strings.Trim(orderParts[1], " ")}
	default:
		return OrderBy{}, errors.New("invalid ordering information")
	}

	return orderBy, nil
}
