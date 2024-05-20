package productbus

import (
	"fmt"
	"regexp"
)

var r = regexp.MustCompile("^[a-zA-Z0-9' -]{3,20}$")

type nameSet struct{}

var Names nameSet

// Parse parses the string value and returns a name if the value complies
// with the rules for a name.
func (nameSet) Parse(value string) (Name, error) {
	if !r.MatchString(value) {
		return Name{}, fmt.Errorf("invalid name %q", value)
	}

	return Name{value}, nil
}

// MustParse parses the string value and returns a name if the value complies
// with the rules for a name. If an error occurs the function panics.
func (nameSet) MustParse(value string) Name {
	name, err := Names.Parse(value)
	if err != nil {
		panic(err)
	}

	return name
}

// =============================================================================

// Name represents a name in the system.
type Name struct {
	name string
}

// String returns the value of the name.
func (n Name) String() string {
	return n.name
}

// UnmarshalText implement the unmarshal interface for JSON conversions.
func (n *Name) UnmarshalText(data []byte) error {
	name, err := Names.Parse(string(data))
	if err != nil {
		return err
	}

	n.name = name.name
	return nil
}

// MarshalText implement the marshal interface for JSON conversions.
func (n Name) MarshalText() ([]byte, error) {
	return []byte(n.name), nil
}

// Equal provides support for the go-cmp package and testing.
func (n Name) Equal(n2 Name) bool {
	return n.name == n2.name
}
