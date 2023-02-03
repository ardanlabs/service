// Package order provides support for describing the ordering of data.
package order

import (
	"errors"
	"fmt"
	"regexp"
)

// Used to check for sql injection problems.
var sqlInjection = regexp.MustCompile("^[A-Za-z0-9_]+$")

// =============================================================================

// Individual directions in the system.
var (
	ASC  = Direction{"ASC"}
	DESC = Direction{"DESC"}
)

// Set of known directions.
var directions = map[string]Direction{
	ASC.name:  ASC,
	DESC.name: DESC,
}

// Direction defines an order direction.
type Direction struct {
	name string
}

// ParseDirection converts a string to a type Direction.
func ParseDirection(value string) (Direction, error) {
	direction, exists := directions[value]
	if !exists {
		return Direction{}, errors.New("invalid direction")
	}

	return direction, nil
}

// Name returns the name of the direction.
func (d Direction) Name() string {
	return d.name
}

// =============================================================================

// Field represents a field of database being managed.
type Field struct {
	name string
}

// ParseField constructs a Field value and checks for potential sql
// injection issues.
func ParseField(field string) (Field, error) {
	if !sqlInjection.MatchString(field) {
		return Field{}, fmt.Errorf("invalid field %q format", field)
	}

	return Field{field}, nil
}

// MustParseField constructs a Field value and checks for potential sql
// injection issues. If there is an error it will panic.
func MustParseField(field string) Field {
	f, err := ParseField(field)
	if err != nil {
		panic(err)
	}

	return f
}

// Name returns the name of the field.
func (f Field) Name() string {
	return f.name
}

// =============================================================================

// FieldSet maintains a set of fields that belong to an entity.
type FieldSet struct {
	fields map[string]Field
}

// NewFieldSet takes a comma delimited set of fields to add to the set.
func NewFieldSet(fields ...Field) FieldSet {
	m := make(map[string]Field)

	for _, field := range fields {
		m[field.Name()] = field
	}

	return FieldSet{
		fields: m,
	}
}

// ParseField takes a field by string and validates it belongs to the set.
// Then returns that field in its proper type.
func (fs FieldSet) ParseField(field string) (Field, error) {
	f, exists := fs.fields[field]
	if !exists {
		return Field{}, fmt.Errorf("field %q not found", field)
	}

	return f, nil
}

// =============================================================================

// By represents a field used to order by and direction.
type By struct {
	field     Field
	direction Direction
}

// NewBy constructs a new By value with no checks.
func NewBy(field Field, direction Direction) By {
	by := By{
		field:     field,
		direction: direction,
	}

	return by
}

// Field returns the field value.
func (b By) Field() Field {
	return b.field
}

// Direction returns the direction value.
func (b By) Direction() Direction {
	return b.direction
}

// Clause returns a sql string with the ordering information.
func (b By) Clause() (string, error) {
	return b.Field().Name() + " " + b.Direction().Name(), nil
}
