// Package auth provides authentication and authorization support.
package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"sync"

	"github.com/golang-jwt/jwt/v4"
)

// ErrForbidden is returned when a auth issue is identified.
var ErrForbidden = errors.New("attempted action is not allowed")

// KeyLookup declares a method set of behavior for looking up
// private and public keys for JWT use.
type KeyLookup interface {
	PrivateKey(kid string) (*rsa.PrivateKey, error)
	PublicKey(kid string) (*rsa.PublicKey, error)
}

// Auth is used to authenticate clients. It can generate a token for a
// set of user claims and recreate the claims by parsing the token.
type Auth struct {
	activeKID string
	keyLookup KeyLookup
	method    jwt.SigningMethod
	parser    *jwt.Parser
	mu        sync.RWMutex
	cache     map[string]*rsa.PublicKey
}

// New creates an Auth to support authentication/authorization.
func New(activeKID string, keyLookup KeyLookup) (*Auth, error) {
	a := Auth{
		activeKID: activeKID,
		keyLookup: keyLookup,
		method:    jwt.GetSigningMethod("RS256"),
		parser:    jwt.NewParser(jwt.WithValidMethods([]string{"RS256"})),
		cache:     make(map[string]*rsa.PublicKey),
	}

	return &a, nil
}

// GenerateToken generates a signed JWT token string representing the user Claims.
func (a *Auth) GenerateToken(claims Claims) (string, error) {
	token := jwt.NewWithClaims(a.method, claims)
	token.Header["kid"] = a.activeKID

	privateKey, err := a.keyLookup.PrivateKey(a.activeKID)
	if err != nil {
		return "", fmt.Errorf("private key: %w", err)
	}

	str, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return str, nil
}

// ValidateToken recreates the Claims that were used to generate a token. It
// verifies that the token was signed using our key.
func (a *Auth) ValidateToken(tokenStr string) (Claims, error) {
	var claims Claims
	token, err := a.parser.ParseWithClaims(tokenStr, &claims, a.publicKeyLookup())
	if err != nil {
		return Claims{}, fmt.Errorf("parsing token: %w", err)
	}

	if !token.Valid {
		return Claims{}, errors.New("invalid token")
	}

	return claims, nil
}

// =============================================================================

// publicKeyLookup implements the JWT key lookup function for returning the public
// key based on the kid in the token header.
func (a *Auth) publicKeyLookup() func(t *jwt.Token) (any, error) {
	f := func(t *jwt.Token) (any, error) {
		kid, ok := t.Header["kid"]
		if !ok {
			return nil, errors.New("missing key id (kid) in token header")
		}

		kidID, ok := kid.(string)
		if !ok {
			return nil, errors.New("user token key id (kid) must be string")
		}

		pubKey, err := func() (*rsa.PublicKey, error) {
			a.mu.RLock()
			defer a.mu.RUnlock()
			pubKey, exists := a.cache[kidID]
			if !exists {
				return nil, errors.New("not found")
			}
			return pubKey, nil
		}()
		if err == nil {
			return pubKey, nil
		}

		pubKey, err = a.keyLookup.PublicKey(kidID)
		if err != nil {
			return nil, fmt.Errorf("fetching public key: %w", err)
		}

		a.mu.Lock()
		defer a.mu.Unlock()
		a.cache[kidID] = pubKey

		return pubKey, nil
	}

	return f
}
