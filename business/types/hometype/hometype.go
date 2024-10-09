// Package hometype represents the home type in the system.
package hometype

import "fmt"

// The set of types that can be used.
var (
	Single = newType("SINGLE FAMILY")
	Condo  = newType("CONDO")
)

// =============================================================================

// Set of known housing types.
var homeTypes = make(map[string]HomeType)

// HomeType represents a type in the system.
type HomeType struct {
	value string
}

func newType(homeType string) HomeType {
	ht := HomeType{homeType}
	homeTypes[homeType] = ht
	return ht
}

// String returns the name of the type.
func (ht HomeType) String() string {
	return ht.value
}

// Equal provides support for the go-cmp package and testing.
func (ht HomeType) Equal(ht2 HomeType) bool {
	return ht.value == ht2.value
}

// MarshalText provides support for logging and any marshal needs.
func (ht HomeType) MarshalText() ([]byte, error) {
	return []byte(ht.value), nil
}

// =============================================================================

// Parse parses the string value and returns a home type if one exists.
func Parse(value string) (HomeType, error) {
	typ, exists := homeTypes[value]
	if !exists {
		return HomeType{}, fmt.Errorf("invalid home type %q", value)
	}

	return typ, nil
}

// MustParse parses the string value and returns a home type if one exists. If
// an error occurs the function panics.
func MustParse(value string) HomeType {
	typ, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return typ
}
