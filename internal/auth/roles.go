package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

// These are the expected values for Claims.Roles.
const (
	RoleAdmin = "ADMIN"
	RoleUser  = "USER"
)

// ctxKey represents the type of value for the context key.
type ctxKey int

// Key is used to store/retrieve a Claims value from a context.Context.
const Key ctxKey = 1

// Claims represents the authorization claims transmitted via a JWT.
type Claims struct {
	Roles []string `json:"roles"`
	jwt.StandardClaims
}

// Valid is called during the parsing of a token.
func (c Claims) Valid() error {
	if err := c.StandardClaims.Valid(); err != nil {
		return errors.Wrap(err, "validating standard claims")
	}

	return nil
}

// HasRole returns true if the claims has at least one of the provided roles.
func (c Claims) HasRole(roles ...string) bool {
	for _, has := range c.Roles {
		for _, want := range roles {
			if has == want {
				return true
			}
		}
	}
	return false
}
