// Package money represents a money in the system.
package money

import (
	"fmt"
)

// Money represents a money in the system.
type Money struct {
	value float64
}

// Value returns the float value of the money.
func (m Money) Value() float64 {
	return m.value
}

// String returns the value of the money.
func (m Money) String() string {
	return fmt.Sprintf("%.2f", m.value)
}

// Equal provides support for the go-cmp package and testing.
func (m Money) Equal(m2 Money) bool {
	return m.value == m2.value
}

// MarshalText provides support for logging and any marshal needs.
func (m Money) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

// =============================================================================

// Parse parses the float value and returns a money if the value complies
// with the rules for money.
func Parse(value float64) (Money, error) {
	if value < 0 || value > 1_000_000 {
		return Money{}, fmt.Errorf("invalid money %.2f", value)
	}

	return Money{value}, nil
}

// MustParse parses the string value and returns a money if the value
// complies with the rules for a money. If an error occurs the function panics.
func MustParse(value float64) Money {
	money, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return money
}
