package auth

import (
	"crypto/rsa"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

// signingAlg is the agreed-upon algorithm that this application uses to sign
// and validate tokens.
//
// The method used to sign the JWT must be validated to avoid a critical
// vulnerability:
// https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
var signingAlg = jwt.SigningMethodRS256

const (
	RoleAdmin = "ADMIN"
	RoleTODO  = "TODO:MORE_ROLES"
)

// KeyFunc is used to map a JWT key id (kid) to the corresponding public key.
type KeyFunc func(kid string) (*rsa.PublicKey, error)

// Claims represents the authorization claims transmitted via a JWT.
type Claims struct {
	Roles []string `json:"roles"`
	jwt.StandardClaims
}

// GenerateToken generates a JWT token string.
func GenerateToken(key *rsa.PrivateKey, kid string, claims Claims) (string, error) {
	tkn := jwt.NewWithClaims(signingAlg, claims)
	tkn.Header["kid"] = kid
	return tkn.SignedString(key)
}

// NewParser is the factory function for a Parser.
func NewParser(kf KeyFunc) *Parser {
	return &Parser{
		kf: kf,
		p: &jwt.Parser{
			ValidMethods: []string{signingAlg.Name},
		},
	}
}

// Parser wraps jwt.Parser with the ability to fetch keys based on kid.
type Parser struct {
	kf KeyFunc
	p  *jwt.Parser
}

// ParseClaims from a token string.
func (p *Parser) ParseClaims(tknStr string) (Claims, error) {
	var claims Claims

	tkn, err := jwt.ParseWithClaims(tknStr, &claims, func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header["kid"]
		if !ok {
			return nil, errors.New("Missing key id (kid) in token header")
		}
		kidStr, ok := kid.(string)
		if !ok {
			return nil, errors.New("Token key id (kid) must be string")
		}

		return p.kf(kidStr)
	})
	if err != nil {
		return Claims{}, err
	}

	if !tkn.Valid {
		return Claims{}, errors.New("Invalid token")
	}

	return claims, nil
}
