// Package name represents a name in the system.
package name

import (
	"fmt"
	"regexp"
)

// Name represents a name in the system.
type Name struct {
	value string
}

// String returns the value of the name.
func (n Name) String() string {
	return n.value
}

// Equal provides support for the go-cmp package and testing.
func (n Name) Equal(n2 Name) bool {
	return n.value == n2.value
}

// =============================================================================

// Null represents a name in the system that can be empty.
type Null struct {
	value string
}

// String returns the value of the name.
func (n Null) String() string {
	return n.value
}

// Equal provides support for the go-cmp package and testing.
func (n Null) Equal(n2 Null) bool {
	return n.value == n2.value
}

// IsNull tests if the value is null.
func (n Null) IsNull() bool {
	return n.value == ""
}

// =============================================================================

var nameRegEx = regexp.MustCompile("^[a-zA-Z0-9' -]{3,20}$")

// Parse parses the string value and returns a name if the value complies
// with the rules for a name.
func Parse(value string) (Name, error) {
	if !nameRegEx.MatchString(value) {
		return Name{}, fmt.Errorf("invalid name %q", value)
	}

	return Name{value}, nil
}

// MustParse parses the string value and returns a name if the value
// complies with the rules for a name. If an error occurs the function panics.
func MustParse(value string) Name {
	name, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return name
}

// ParseNull parses the string value and returns a name if the value complies
// with the rules for a name.
func ParseNull(value string) (Null, error) {
	if value == "" {
		return Null{}, nil
	}

	if !nameRegEx.MatchString(value) {
		return Null{}, fmt.Errorf("invalid name %q", value)
	}

	return Null{value}, nil
}

// MustParseNull parses the string value and returns a name if the value
// complies with the rules for a name. If an error occurs the function panics.
func MustParseNull(value string) Null {
	name, err := ParseNull(value)
	if err != nil {
		panic(err)
	}

	return name
}
