package user

import (
	"context"
	"database/sql"
	"time"

	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/bcrypt"
)

const usersCollection = "users"

var (
	// ErrNotFound is used when a specific User is requested but does not exist.
	ErrNotFound = errors.New("User not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrAuthenticationFailure occurs when a user attempts to authenticate but
	// anything goes wrong.
	ErrAuthenticationFailure = errors.New("Authentication failed")

	// ErrForbidden occurs when a user tries to do something that is forbidden to them according to our access control policies.
	ErrForbidden = errors.New("Attempted action is not allowed")
)

// List retrieves a list of existing users from the database.
func List(ctx context.Context, db *sqlx.DB) ([]User, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.List")
	defer span.End()

	users := []User{}
	const q = `SELECT * FROM users`

	if err := db.SelectContext(ctx, &users, q); err != nil {
		return nil, errors.Wrap(err, "selecting users")
	}

	return users, nil
}

// Retrieve gets the specified user from the database.
func Retrieve(ctx context.Context, claims auth.Claims, db *sqlx.DB, id string) (*User, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	// If you are not an admin and looking to retrieve someone else then you are rejected.
	if !claims.HasRole(auth.RoleAdmin) && claims.Subject != id {
		return nil, ErrForbidden
	}

	var u User
	const q = `SELECT * FROM users WHERE user_id = $1`
	if err := db.GetContext(ctx, &u, q, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting user %q", id)
	}

	return &u, nil
}

// Create inserts a new user into the database.
func Create(ctx context.Context, db *sqlx.DB, n *NewUser, now time.Time) (*User, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.Create")
	defer span.End()

	hash, err := bcrypt.GenerateFromPassword([]byte(n.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, "generating password hash")
	}

	u := User{
		ID:           uuid.New().String(),
		Name:         n.Name,
		Email:        n.Email,
		PasswordHash: hash,
		Roles:        n.Roles,
		DateCreated:  now.UTC(),
		DateUpdated:  now.UTC(),
	}

	const q = `INSERT INTO users
		(user_id, name, email, password_hash, roles, date_created, date_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = db.ExecContext(
		ctx, q,
		u.ID, u.Name, u.Email,
		u.PasswordHash, u.Roles,
		u.DateCreated, u.DateUpdated,
	)
	if err != nil {
		return nil, errors.Wrap(err, "inserting user")
	}

	return &u, nil
}

// Update replaces a user document in the database.
func Update(ctx context.Context, claims auth.Claims, db *sqlx.DB, id string, upd *UpdateUser, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.user.Update")
	defer span.End()

	u, err := Retrieve(ctx, claims, db, id)
	if err != nil {
		return err
	}

	if upd.Name != nil {
		u.Name = *upd.Name
	}
	if upd.Email != nil {
		u.Email = *upd.Email
	}
	if upd.Roles != nil {
		u.Roles = upd.Roles
	}
	if upd.Password != nil {
		pw, err := bcrypt.GenerateFromPassword([]byte(*upd.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.Wrap(err, "generating password hash")
		}
		u.PasswordHash = pw
	}

	u.DateUpdated = now

	const q = `UPDATE users SET
		"name" = $2,
		"email" = $3,
		"roles" = $4,
		"password_hash" = $5,
		"date_updated" = $6
		WHERE user_id = $1`
	_, err = db.ExecContext(ctx, q, id,
		u.Name, u.Email, u.Roles,
		u.PasswordHash, u.DateUpdated,
	)
	if err != nil {
		return errors.Wrap(err, "updating user")
	}

	return nil
}

// Delete removes a user from the database.
func Delete(ctx context.Context, db *sqlx.DB, id string) error {
	ctx, span := trace.StartSpan(ctx, "internal.user.Delete")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}

	const q = `DELETE FROM users WHERE user_id = $1`

	if _, err := db.ExecContext(ctx, q, id); err != nil {
		return errors.Wrapf(err, "deleting user %s", id)
	}

	return nil
}

// Authenticate finds a user by their email and verifies their password. On
// success it returns a Claims value representing this user. The claims can be
// used to generate a token for future authentication.
func Authenticate(ctx context.Context, db *sqlx.DB, now time.Time, email, password string) (auth.Claims, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.Authenticate")
	defer span.End()

	const q = `SELECT * FROM users WHERE email = $1`

	var u User
	if err := db.GetContext(ctx, &u, q, email); err != nil {

		// Normally we would return ErrNotFound in this scenario but we do not want
		// to leak to an unauthenticated user which emails are in the system.
		if err == sql.ErrNoRows {
			return auth.Claims{}, ErrAuthenticationFailure
		}

		return auth.Claims{}, errors.Wrap(err, "selecting single user")
	}

	// Compare the provided password with the saved hash. Use the bcrypt
	// comparison function so it is cryptographically secure.
	if err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password)); err != nil {
		return auth.Claims{}, ErrAuthenticationFailure
	}

	// If we are this far the request is valid. Create some claims for the user
	// and generate their token.
	claims := auth.NewClaims(u.ID, u.Roles, now, time.Hour)
	return claims, nil
}
