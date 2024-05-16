package homebus

import "fmt"

type typeSet struct {
	Single Type
	Condo  Type
}

// Types represents the set of types that can be used.
var Types = typeSet{
	Single: newType("SINGLE FAMILY"),
	Condo:  newType("CONDO"),
}

// Parse parses the string value and returns a type if one exists.
func (typeSet) Parse(value string) (Type, error) {
	typ, exists := types[value]
	if !exists {
		return Type{}, fmt.Errorf("invalid type %q", value)
	}

	return typ, nil
}

// MustParse parses the string value and returns a type if one exists. If
// an error occurs the function panics.
func (typeSet) MustParse(value string) Type {
	typ, err := Types.Parse(value)
	if err != nil {
		panic(err)
	}

	return typ
}

// =============================================================================

// Set of known housing types.
var types = make(map[string]Type)

// Type represents a type in the system.
type Type struct {
	name string
}

func newType(typ string) Type {
	t := Type{typ}
	types[typ] = t
	return t
}

// String returns the name of the type.
func (t Type) String() string {
	return t.name
}

// UnmarshalText implement the unmarshal interface for JSON conversions.
func (t *Type) UnmarshalText(data []byte) error {
	typ, err := Types.Parse(string(data))
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
