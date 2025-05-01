// Package domain represents the domain type in the system.
package domain

import "fmt"

// The set of roles that can be used.
var (
	User    = newDomain("USER")
	Product = newDomain("PRODUCT")
	Home    = newDomain("HOME")
)

// =============================================================================

// Set of known domains.
var domains = make(map[string]Domain)

// Domain represents a domain in the system.
type Domain struct {
	value string
}

func newDomain(domain string) Domain {
	d := Domain{domain}
	domains[domain] = d
	return d
}

// String returns the name of the role.
func (d Domain) String() string {
	return d.value
}

// Equal provides support for the go-cmp package and testing.
func (d Domain) Equal(d2 Domain) bool {
	return d.value == d2.value
}

// MarshalText provides support for logging and any marshal needs.
func (d Domain) MarshalText() ([]byte, error) {
	return []byte(d.value), nil
}

// =============================================================================

// Parse parses the string value and returns a role if one exists.
func Parse(value string) (Domain, error) {
	domain, exists := domains[value]
	if !exists {
		return Domain{}, fmt.Errorf("invalid domain %q", value)
	}

	return domain, nil
}

// MustParse parses the string value and returns a role if one exists. If
// an error occurs the function panics.
func MustParse(value string) Domain {
	domain, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return domain
}
