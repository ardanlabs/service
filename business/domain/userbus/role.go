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

// ToStringSlice takes a collection of user roles and converts them to a slice
// of string.
func (roleSet) ToStringSlice(usrRoles []Role) []string {
	roles := make([]string, len(usrRoles))
	for i, role := range usrRoles {
		roles[i] = role.String()
	}

	return roles
}

// ToRoleSlice takes a collection of strings and converts them to a slice
// of roles.
func (roleSet) ToRoleSlice(roles []string) ([]Role, error) {
	usrRoles := make([]Role, len(roles))
	for i, roleStr := range roles {
		role, err := Roles.Parse(roleStr)
		if err != nil {
			return nil, err
		}
		usrRoles[i] = role
	}

	return usrRoles, nil
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

// Equal provides support for the go-cmp package and testing.
func (r Role) Equal(r2 Role) bool {
	return r.name == r2.name
}
