package auth

import (
	"crypto/rsa"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
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

// NewClaims constructs a Claims value for the identified user. The Claims
// expire within a specified duration of the provided time. Additional fields
// of the Claims can be set after calling NewClaims is desired.
func NewClaims(subject string, roles []string, now time.Time, expires time.Duration) Claims {
	c := Claims{
		Roles: roles,
		StandardClaims: jwt.StandardClaims{
			Subject:   subject,
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(expires).Unix(),
		},
	}

	return c
}

// Valid is called during the parsing of a token.
func (c Claims) Valid() error {
	for _, r := range c.Roles {
		switch r {
		case RoleAdmin, RoleUser: // Role is valid.
		default:
			return fmt.Errorf("invalid role %q", r)
		}
	}
	if err := c.StandardClaims.Valid(); err != nil {
		return errors.Wrap(err, "validating standard claims")
	}
	return nil
}

// GenerateToken generates a signed JWT token string.
func GenerateToken(key *rsa.PrivateKey, keyID string, alg string, claims Claims) (string, error) {
	method := jwt.GetSigningMethod(alg)
	if method == nil {
		return "", errors.Errorf("unknown algorithm %v", alg)
	}

	tkn := jwt.NewWithClaims(method, claims)
	tkn.Header["kid"] = keyID
	str, err := tkn.SignedString(key)
	if err != nil {
		return "", errors.Wrap(err, "signing token")
	}

	return str, nil
}

// KeyFunc is used to map a JWT key id (kid) to the corresponding public key.
// * Private keys should be rotated. During the transition
// period, tokens signed with the old and new keys can coexist by looking up
// the correct public key by key id (kid).
// * Key-id-to-public-key resolution is usually accomplished via a public JWKS
// endpoint. See https://auth0.com/docs/jwks for more details.
type KeyFunc func(keyID string) (*rsa.PublicKey, error)

// Parser wraps jwt.Parser with the ability to fetch keys based on kid.
type Parser struct {
	kf KeyFunc
	p  *jwt.Parser
}

// NewParser is the factory function for a Parser.
func NewParser(kf KeyFunc, validAlgNames []string) *Parser {
	return &Parser{
		kf: kf,
		p: &jwt.Parser{

			// The algorithm used to sign the JWT must be validated to avoid a critical
			// vulnerability:
			// https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
			ValidMethods: validAlgNames,
		},
	}
}

// ParseClaims from a token string.
func (p *Parser) ParseClaims(tknStr string) (Claims, error) {

	// Use the parsed (but unverified token) to resolve the correct key to do
	// verification with.
	f := func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header["kid"]
		if !ok {
			return nil, errors.New("Missing key id (kid) in token header")
		}
		kidStr, ok := kid.(string)
		if !ok {
			return nil, errors.New("Token key id (kid) must be string")
		}

		return p.kf(kidStr)
	}

	var claims Claims
	tkn, err := jwt.ParseWithClaims(tknStr, &claims, f)
	if err != nil {
		return Claims{}, errors.Wrap(err, "parsing token")
	}

	if !tkn.Valid {
		return Claims{}, errors.New("Invalid token")
	}

	return claims, nil
}
