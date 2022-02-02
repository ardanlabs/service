// Package user provides an example of a core business API. Right now these
// calls are just wrapping the data/data layer. But at some point you will
// want auditing or something that isn't specific to the data/store layer.
package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/core/user/db"
	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/ardanlabs/service/business/sys/database"
	"github.com/ardanlabs/service/business/sys/validate"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("user not found")
	ErrInvalidID             = errors.New("ID is not in its proper form")
	ErrInvalidEmail          = errors.New("email is not valid")
	ErrUniqueEmail           = errors.New("email is not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

// Core manages the set of APIs for user access.
type Core struct {
	store db.Store
}

// NewCore constructs a core for user api access.
func NewCore(log *zap.SugaredLogger, sqlxDB *sqlx.DB) Core {
	return Core{
		store: db.NewStore(log, sqlxDB),
	}
}

// Create inserts a new user into the database.
func (c Core) Create(ctx context.Context, nu NewUser, now time.Time) (User, error) {
	if err := validate.Check(nu); err != nil {
		return User{}, fmt.Errorf("validating data: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("generating password hash: %w", err)
	}

	dbUsr := db.User{
		ID:           validate.GenerateID(),
		Name:         nu.Name,
		Email:        nu.Email,
		PasswordHash: hash,
		Roles:        nu.Roles,
		DateCreated:  now,
		DateUpdated:  now,
	}

	// This provides an example of how to execute a transaction if required.
	tran := func(tx sqlx.ExtContext) error {
		if err := c.store.Tran(tx).Create(ctx, dbUsr); err != nil {
			if errors.Is(err, database.ErrDBDuplicatedEntry) {
				return fmt.Errorf("create: %w", ErrUniqueEmail)
			}
			return fmt.Errorf("create: %w", err)
		}
		return nil
	}

	if err := c.store.WithinTran(ctx, tran); err != nil {
		return User{}, fmt.Errorf("tran: %w", err)
	}

	return toUser(dbUsr), nil
}

// Update replaces a user document in the database.
func (c Core) Update(ctx context.Context, userID string, uu UpdateUser, now time.Time) error {
	if err := validate.CheckID(userID); err != nil {
		return ErrInvalidID
	}

	if err := validate.Check(uu); err != nil {
		return fmt.Errorf("validating data: %w", err)
	}

	dbUsr, err := c.store.QueryByID(ctx, userID)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("updating user userID[%s]: %w", userID, err)
	}

	if uu.Name != nil {
		dbUsr.Name = *uu.Name
	}
	if uu.Email != nil {
		dbUsr.Email = *uu.Email
	}
	if uu.Roles != nil {
		dbUsr.Roles = uu.Roles
	}
	if uu.Password != nil {
		pw, err := bcrypt.GenerateFromPassword([]byte(*uu.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("generating password hash: %w", err)
		}
		dbUsr.PasswordHash = pw
	}
	dbUsr.DateUpdated = now

	if err := c.store.Update(ctx, dbUsr); err != nil {
		if errors.Is(err, database.ErrDBDuplicatedEntry) {
			return fmt.Errorf("updating user userID[%s]: %w", userID, ErrUniqueEmail)
		}
		return fmt.Errorf("update: %w", err)
	}

	return nil
}

// Delete removes a user from the database.
func (c Core) Delete(ctx context.Context, userID string) error {
	if err := validate.CheckID(userID); err != nil {
		return ErrInvalidID
	}

	if err := c.store.Delete(ctx, userID); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of existing users from the database.
func (c Core) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]User, error) {
	dbUsers, err := c.store.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return toUserSlice(dbUsers), nil
}

// QueryByID gets the specified user from the database.
func (c Core) QueryByID(ctx context.Context, userID string) (User, error) {
	if err := validate.CheckID(userID); err != nil {
		return User{}, ErrInvalidID
	}

	dbUsr, err := c.store.QueryByID(ctx, userID)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("query: %w", err)
	}

	return toUser(dbUsr), nil
}

// QueryByEmail gets the specified user from the database by email.
func (c Core) QueryByEmail(ctx context.Context, email string) (User, error) {

	// Email Validate function in validate.
	if !validate.CheckEmail(email) {
		return User{}, ErrInvalidEmail
	}

	dbUsr, err := c.store.QueryByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("query: %w", err)
	}

	return toUser(dbUsr), nil
}

// Authenticate finds a user by their email and verifies their password. On
// success it returns a Claims User representing this user. The claims can be
// used to generate a token for future authentication.
func (c Core) Authenticate(ctx context.Context, now time.Time, email, password string) (auth.Claims, error) {
	dbUsr, err := c.store.QueryByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return auth.Claims{}, ErrNotFound
		}
		return auth.Claims{}, fmt.Errorf("query: %w", err)
	}

	// Compare the provided password with the saved hash. Use the bcrypt
	// comparison function so it is cryptographically secure.
	if err := bcrypt.CompareHashAndPassword(dbUsr.PasswordHash, []byte(password)); err != nil {
		return auth.Claims{}, ErrAuthenticationFailure
	}

	// If we are this far the request is valid. Create some claims for the user
	// and generate their token.
	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   dbUsr.ID,
			Issuer:    "service project",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: dbUsr.Roles,
	}

	return claims, nil
}
