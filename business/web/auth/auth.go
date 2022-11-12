// Package auth provides authentication and authorization support.
package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
	"github.com/open-policy-agent/opa/rego"
	"go.uber.org/zap"
)

// ErrForbidden is returned when a auth issue is identified.
var ErrForbidden = errors.New("attempted action is not allowed")

// KeyLookup declares a method set of behavior for looking up
// private and public keys for JWT use.
type KeyLookup interface {
	PrivateKeyPEM(kid string) (string, error)
	PublicKeyPEM(kid string) (string, error)
}

// Config represents information required to initialize auth.
type Config struct {
	Log       *zap.SugaredLogger
	DB        *sqlx.DB
	KeyLookup KeyLookup
}

// Auth is used to authenticate clients. It can generate a token for a
// set of user claims and recreate the claims by parsing the token.
type Auth struct {
	log       *zap.SugaredLogger
	db        *sqlx.DB
	keyLookup KeyLookup
	method    jwt.SigningMethod
	parser    *jwt.Parser
	mu        sync.RWMutex
	cache     map[string]*rsa.PublicKey
}

// New creates an Auth to support authentication/authorization.
func New(cfg Config) (*Auth, error) {
	a := Auth{
		log:       cfg.Log,
		db:        cfg.DB,
		keyLookup: cfg.KeyLookup,
		method:    jwt.GetSigningMethod("RS256"),
		parser:    jwt.NewParser(jwt.WithValidMethods([]string{"RS256"})),
		cache:     make(map[string]*rsa.PublicKey),
	}

	return &a, nil
}

// GenerateToken generates a signed JWT token string representing the user Claims.
func (a *Auth) GenerateToken(kid string, claims Claims) (string, error) {
	token := jwt.NewWithClaims(a.method, claims)
	token.Header["kid"] = kid

	privateKeyPEM, err := a.keyLookup.PrivateKeyPEM(kid)
	if err != nil {
		return "", fmt.Errorf("private key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return "", fmt.Errorf("parsing private pem: %w", err)
	}

	str, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return str, nil
}

// Authenticate processes the token to validate the sender's token is valid.
func (a *Auth) Authenticate(ctx context.Context, bearerToken string) (Claims, error) {
	parts := strings.Split(bearerToken, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return Claims{}, errors.New("expected authorization header format: Bearer <token>")
	}

	var claims Claims
	token, err := a.parser.ParseWithClaims(parts[1], &claims, a.publicKeyLookup())
	if err != nil {
		return Claims{}, fmt.Errorf("error parsing token: %w", err)
	}

	if !token.Valid {
		return Claims{}, errors.New("token failed signature check")
	}

	// Perform an extra level of authentication verification with OPA.

	pem, err := a.keyLookup.PublicKeyPEM(token.Header["kid"].(string))
	if err != nil {
		return Claims{}, fmt.Errorf("failed to fetch public key: %w", err)
	}

	input := map[string]any{
		"Key":   pem,
		"Token": parts[1],
	}

	if err := a.opaPolicyEvaluation(ctx, opaAuthentication, input); err != nil {
		return Claims{}, fmt.Errorf("authentication failed : %w", err)
	}

	// Check the database for this user to verify they are still enabled.

	if !a.isUserEnabled(ctx, claims) {
		return Claims{}, fmt.Errorf("user not enabled : %w", err)
	}

	return claims, nil
}

// Authorize attempts to authorize the user with the provided input roles, if
// none of the input roles are within the user's claims, we return an error
// otherwise the user is authorized.
func (a *Auth) Authorize(ctx context.Context, claims Claims, roles ...string) error {

	// If no roles are provided, default to using all known roles to authorize.
	if len(roles) == 0 {
		roles = []string{RoleAdmin, RoleUser}
	}

	input := map[string]any{
		"Roles":      claims.Roles,
		"InputRoles": roles,
	}

	if err := a.opaPolicyEvaluation(ctx, opaAuthorization, input); err != nil {
		return fmt.Errorf("rego evaluation failed : %w", err)
	}

	return nil
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

		pem, err := a.keyLookup.PublicKeyPEM(kidID)
		if err != nil {
			return nil, fmt.Errorf("fetching public key: %w", err)
		}

		pubKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte(pem))
		if err != nil {
			return nil, fmt.Errorf("parsing public pem: %w", err)
		}

		a.mu.Lock()
		defer a.mu.Unlock()
		a.cache[kidID] = pubKey

		return pubKey, nil
	}

	return f
}

// opaPolicyEvaluation asks opa to evaulate the token against the specified token
// policy and public key.
func (a *Auth) opaPolicyEvaluation(ctx context.Context, opaPolicy string, input any) error {
	query, err := rego.New(
		rego.Query(fmt.Sprintf("x = data.%s.allow", opaPackage)),
		rego.Module("policy.rego", opaPolicy),
	).PrepareForEval(ctx)
	if err != nil {
		return err
	}

	results, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	if len(results) == 0 {
		return errors.New("no results")
	}

	result, ok := results[0].Bindings["x"].(bool)
	if !ok || !result {
		return fmt.Errorf("bindings results[%v] ok[%v]", results, ok)
	}

	return nil
}

// isUserEnabled hits the database and checks the user is not disabled. If the
// user is not found in the database, they are still considered validated by
// their token.
func (a *Auth) isUserEnabled(ctx context.Context, claims Claims) bool {

	// For now until I learn more, not having a db connection means the
	// account is trused.
	// if a.db == nil {
	// 	return true
	// }

	// data := struct {
	// 	KeycloakId string `db:"keycloak_id"`
	// }{
	// 	KeycloakId: claims.Subject,
	// }

	// const query = `
	// SELECT
	// 	id
	// FROM
	// 	users
	// WHERE
	// 	user_id = :user_id AND
	// 	enabled = true,
	// LIMIT 1;`

	// var usr user.User
	// if err := database.NamedQueryStruct(ctx, a.log, a.db, query, data, &usr); err != nil {
	// 	return err == database.ErrDBNotFound
	// }

	// if !usr.Enabled {
	// 	return false
	// }

	return true
}
