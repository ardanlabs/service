// Package auth provides authentication and authorization support.
package auth

import (
	"crypto/rsa"

	"github.com/dgrijalva/jwt-go/v4"
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
	jwt.StandardClaims
	Roles []string `json:"roles"`
}

// Authorized returns true if the claims has at least one of the provided roles.
func (c Claims) Authorized(roles ...string) bool {
	for _, has := range c.Roles {
		for _, want := range roles {
			if has == want {
				return true
			}
		}
	}
	return false
}

// KeyLookup declares a method set of behavior for looking up
// private and public keys for JWT use.
type KeyLookup interface {
	PrivateKey(kid string) (*rsa.PrivateKey, error)
	PublicKey(kid string) (*rsa.PublicKey, error)
}

// Auth is used to authenticate clients. It can generate a token for a
// set of user claims and recreate the claims by parsing the token.
type Auth struct {
	algorithm string
	keyLookup KeyLookup
	method    jwt.SigningMethod
	keyFunc   func(t *jwt.Token) (interface{}, error)
	parser    *jwt.Parser
}

// New creates an Auth to support authentication/authorization.
func New(algorithm string, keyLookup KeyLookup) (*Auth, error) {
	method := jwt.GetSigningMethod(algorithm)
	if method == nil {
		return nil, errors.Errorf("unknown algorithm %v", algorithm)
	}

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header["kid"]
		if !ok {
			return nil, errors.New("missing key id (kid) in token header")
		}
		kidID, ok := kid.(string)
		if !ok {
			return nil, errors.New("user token key id (kid) must be string")
		}
		return keyLookup.PublicKey(kidID)
	}

	// Create the token parser to use. The algorithm used to sign the JWT must be
	// validated to avoid a critical vulnerability:
	// https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
	parser := jwt.NewParser(jwt.WithValidMethods([]string{algorithm}), jwt.WithAudience("student"))

	a := Auth{
		algorithm: algorithm,
		keyLookup: keyLookup,
		method:    method,
		keyFunc:   keyFunc,
		parser:    parser,
	}

	return &a, nil
}

// GenerateToken generates a signed JWT token string representing the user Claims.
func (a *Auth) GenerateToken(kid string, claims Claims) (string, error) {
	token := jwt.NewWithClaims(a.method, claims)
	token.Header["kid"] = kid

	privateKey, err := a.keyLookup.PrivateKey(kid)
	if err != nil {
		return "", errors.New("kid lookup failed")
	}

	str, err := token.SignedString(privateKey)
	if err != nil {
		return "", errors.Wrap(err, "signing token")
	}

	return str, nil
}

// ValidateToken recreates the Claims that were used to generate a token. It
// verifies that the token was signed using our key.
func (a *Auth) ValidateToken(tokenStr string) (Claims, error) {
	var claims Claims
	token, err := a.parser.ParseWithClaims(tokenStr, &claims, a.keyFunc)
	if err != nil {
		return Claims{}, errors.Wrap(err, "parsing token")
	}

	if !token.Valid {
		return Claims{}, errors.New("invalid token")
	}

	return claims, nil
}
