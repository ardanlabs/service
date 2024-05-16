package userbus

import "fmt"

type roleSet struct {
	Admin Role
	User  Role
}

// Roles represents the set of roles that can be used.
var Roles = roleSet{
	Admin: newRole("ADMIN"),
	User:  newRole("USER"),
}

// Parse parses the string value and returns a role if one exists.
func (roleSet) Parse(value string) (Role, error) {
	role, exists := roles[value]
	if !exists {
		return Role{}, fmt.Errorf("invalid role %q", value)
	}

	return role, nil
}

// MustParse parses the string value and returns a role if one exists. If
// an error occurs the function panics.
func (roleSet) MustParse(value string) Role {
	role, err := Roles.Parse(value)
	if err != nil {
		panic(err)
	}

	return role
}

// =============================================================================

// Set of known roles.
var roles = make(map[string]Role)

// Role represents a role in the system.
type Role struct {
	name string
}

func newRole(role string) Role {
	r := Role{role}
	roles[role] = r
	return r
}

// String returns the name of the role.
func (r Role) String() string {
	return r.name
}

// UnmarshalText implement the unmarshal interface for JSON conversions.
func (r *Role) UnmarshalText(data []byte) error {
	role, err := Roles.Parse(string(data))
	if err != nil {
		return err
	}

	r.name = role.name
	return nil
}

// MarshalText implement the marshal interface for JSON conversions.
func (r Role) MarshalText() ([]byte, error) {
	return []byte(r.name), nil
}

// Equal provides support for the go-cmp package and testing.
func (r Role) Equal(r2 Role) bool {
	return r.name == r2.name
}
