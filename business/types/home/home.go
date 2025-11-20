// Package home represents the home type in the system.
package home

import "fmt"

// The set of types that can be used.
var (
	Single = newType("SINGLE FAMILY")
	Condo  = newType("CONDO")
)

// =============================================================================

// Set of known housing types.
var homes = make(map[string]Home)

// Home represents a type in the system.
type Home struct {
	value string
}

func newType(homeType string) Home {
	h := Home{homeType}
	homes[homeType] = h
	return h
}

// String returns the name of the type.
func (h Home) String() string {
	return h.value
}

// Equal provides support for the go-cmp package and testing.
func (h Home) Equal(h2 Home) bool {
	return h.value == h2.value
}

// MarshalText provides support for logging and any marshal needs.
func (h Home) MarshalText() ([]byte, error) {
	return []byte(h.value), nil
}

// =============================================================================

// Parse parses the string value and returns a home type if one exists.
func Parse(value string) (Home, error) {
	home, exists := homes[value]
	if !exists {
		return Home{}, fmt.Errorf("invalid home type %q", value)
	}

	return home, nil
}

// MustParse parses the string value and returns a home type if one exists. If
// an error occurs the function panics.
func MustParse(value string) Home {
	typ, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return typ
}
