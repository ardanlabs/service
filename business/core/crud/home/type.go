package home

import "fmt"

// Set of possible roles for a housing type.
var (
	TypeSingle = Type{"SINGLE FAMILY"}
	TypeCondo  = Type{"CONDO"}
)

// Set of known housing types.
var types = map[string]Type{
	TypeSingle.name: TypeSingle,
	TypeCondo.name:  TypeCondo,
}

// Type represents a type in the system.
type Type struct {
	name string
}

// ParseType parses the string value and returns a type if one exists.
func ParseType(value string) (Type, error) {
	typ, exists := types[value]
	if !exists {
		return Type{}, fmt.Errorf("invalid type %q", value)
	}

	return typ, nil
}

// MustParseType parses the string value and returns a type if one exists. If
// an error occurs the function panics.
func MustParseType(value string) Type {
	typ, err := ParseType(value)
	if err != nil {
		panic(err)
	}

	return typ
}

// Name returns the name of the type.
func (t Type) Name() string {
	return t.name
}

// UnmarshalText implement the unmarshal interface for JSON conversions.
func (t *Type) UnmarshalText(data []byte) error {
	typ, err := ParseType(string(data))
	if err != nil {
		return err
	}

	t.name = typ.name
	return nil
}

// MarshalText implement the marshal interface for JSON conversions.
func (t Type) MarshalText() ([]byte, error) {
	return []byte(t.name), nil
}

// Equal provides support for the go-cmp package and testing.
func (t Type) Equal(t2 Type) bool {
	return t.name == t2.name
}
