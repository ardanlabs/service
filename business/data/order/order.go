// Package order provides support for describing the ordering of data.
package order

import (
	"fmt"
	"regexp"
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

// Used to check for sql injection problems.
var valid = regexp.MustCompile("^[A-Za-z0-9_]+$")

// Validate checks the field and direction string for sql injection issues.
func Validate(field string, direction string) error {
	if !valid.MatchString(field) {
		return fmt.Errorf("invalid field %q format", field)
	}

	if !valid.MatchString(direction) {
		return fmt.Errorf("invalid direction %q format", direction)
	}

	if _, exists := directions[direction]; !exists {
		return fmt.Errorf("invalid direction %q format", direction)
	}

	return nil
}

// =============================================================================

// By represents a field used to order by and direction.
type By struct {
	Field     string
	Direction string
}

// NewBy constructs a new By value with no checks.
func NewBy(field string, direction string) By {
	return By{
		Field:     field,
		Direction: direction,
	}
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
	if err := Validate(by.Field, by.Direction); err != nil {
		return err
	}

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
	if err := Validate(field, direction); err != nil {
		return By{}, err
	}

	if _, exists := o.Fields[field]; !exists {
		return By{}, fmt.Errorf("invalid field %q provided", field)
	}

	if _, exists := directions[direction]; !exists {
		return By{}, fmt.Errorf("invalid direction %q provided", direction)
	}

	return By{Field: field, Direction: direction}, nil
}
