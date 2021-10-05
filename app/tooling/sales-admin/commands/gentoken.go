package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/ardanlabs/service/business/sys/database"
	"github.com/ardanlabs/service/foundation/keystore"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

// GenToken generates a JWT for the specified user.
func GenToken(log *zap.SugaredLogger, cfg database.Config, userID string, kid string) error {
	if userID == "" || kid == "" {
		fmt.Println("help: gentoken <user_id> <kid>")
		return ErrHelp
	}

	db, err := database.Open(cfg)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := user.NewCore(log, db)

	usr, err := user.QueryByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("retrieve user: %w", err)
	}

	// Construct a key store based on the key files stored in
	// the specified directory.
	keysFolder := "zarf/keys/"
	ks, err := keystore.NewFS(os.DirFS(keysFolder))
	if err != nil {
		return fmt.Errorf("reading keys: %w", err)
	}

	// Init the auth package.
	activeKID := "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"
	a, err := auth.New(activeKID, ks)
	if err != nil {
		return fmt.Errorf("constructing auth: %w", err)
	}

	// Generating a token requires defining a set of claims. In this applications
	// case, we only care about defining the subject and the user in question and
	// the roles they have on the database. This token will expire in a year.
	//
	// iss (issuer): Issuer of the JWT
	// sub (subject): Subject of the JWT (the user)
	// aud (audience): Recipient for which the JWT is intended
	// exp (expiration time): Time after which the JWT expires
	// nbf (not before time): Time before which the JWT must not be accepted for processing
	// iat (issued at time): Time at which the JWT was issued; can be used to determine age of the JWT
	// jti (JWT ID): Unique identifier; can be used to prevent the JWT from being replayed (allows a token to be used only once)
	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   usr.ID,
			Issuer:    "service project",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8760 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: usr.Roles,
	}

	// This will generate a JWT with the claims embedded in them. The database
	// with need to be configured with the information found in the public key
	// file to validate these claims. Dgraph does not support key rotate at
	// this time.
	token, err := a.GenerateToken(claims)
	if err != nil {
		return fmt.Errorf("generating token: %w", err)
	}

	fmt.Printf("-----BEGIN TOKEN-----\n%s\n-----END TOKEN-----\n", token)
	return nil
}
