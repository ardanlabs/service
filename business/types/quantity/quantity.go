// Package quantity represents a quantity in the system.
package quantity

import (
	"fmt"
)

// Quantity represents a quantity in the system.
type Quantity struct {
	value int
}

// Value returns the int value of the quantity.
func (q Quantity) Value() int {
	return q.value
}

// String returns the value of the quantity.
func (q Quantity) String() string {
	return fmt.Sprintf("%d", q.value)
}

// Equal provides support for the go-cmp package and testing.
func (q Quantity) Equal(q2 Quantity) bool {
	return q.value == q2.value
}

// MarshalText provides support for logging and any marshal needs.
func (q Quantity) MarshalText() ([]byte, error) {
	return []byte(q.String()), nil
}

// =============================================================================

// Parse parses the float value and returns a quantity if the value complies
// with the rules for quantity.
func Parse(value int) (Quantity, error) {
	if value < 0 || value > 1_000_000 {
		return Quantity{}, fmt.Errorf("invalid quantity %d", value)
	}

	return Quantity{value}, nil
}

// MustParse parses the string value and returns a quantity if the value
// complies with the rules for a quantity. If an error occurs the function panics.
func MustParse(value int) Quantity {
	money, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return money
}
