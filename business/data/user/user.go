// Package user contains user related CRUD functionality.
package user

import (
	"context"
	"log"
	"time"

	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/ardanlabs/service/business/sys/validate"
	"github.com/ardanlabs/service/foundation/database"
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

// Store manages the set of API's for user access.
type Store struct {
	log *log.Logger
	db  *sqlx.DB
}

// NewStore constructs a user store for api access.
func NewStore(log *log.Logger, db *sqlx.DB) Store {
	return Store{
		log: log,
		db:  db,
	}
}

// Create inserts a new user into the database.
func (s Store) Create(ctx context.Context, traceID string, nu NewUser, now time.Time) (User, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.create")
	defer span.End()

	if err := validate.Check(nu); err != nil {
		return User{}, errors.Wrap(err, "validating data")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, errors.Wrap(err, "generating password hash")
	}

	usr := User{
		ID:           validate.GenerateID(),
		Name:         nu.Name,
		Email:        nu.Email,
		PasswordHash: hash,
		Roles:        nu.Roles,
		DateCreated:  now.UTC(),
		DateUpdated:  now.UTC(),
	}

	const q = `
	INSERT INTO users
		(user_id, name, email, password_hash, roles, date_created, date_updated)
	VALUES
		(:user_id, :name, :email, :password_hash, :roles, :date_created, :date_updated)`

	s.log.Printf("%s: %s: %s", traceID, "user.Create",
		database.Log(q, usr),
	)

	if _, err := s.db.NamedExecContext(ctx, q, usr); err != nil {
		return User{}, errors.Wrap(err, "inserting user")
	}

	return usr, nil
}

// Update replaces a user document in the database.
func (s Store) Update(ctx context.Context, traceID string, claims auth.Claims, userID string, uu UpdateUser, now time.Time) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.update")
	defer span.End()

	if err := validate.CheckID(userID); err != nil {
		return database.ErrInvalidID
	}
	if err := validate.Check(uu); err != nil {
		return errors.Wrap(err, "validating data")
	}

	usr, err := s.QueryByID(ctx, traceID, claims, userID)
	if err != nil {
		return errors.Wrap(err, "updating user")
	}

	if uu.Name != nil {
		usr.Name = *uu.Name
	}
	if uu.Email != nil {
		usr.Email = *uu.Email
	}
	if uu.Roles != nil {
		usr.Roles = uu.Roles
	}
	if uu.Password != nil {
		pw, err := bcrypt.GenerateFromPassword([]byte(*uu.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.Wrap(err, "generating password hash")
		}
		usr.PasswordHash = pw
	}
	usr.DateUpdated = now

	const q = `
	UPDATE
		users
	SET 
		"name" = :name,
		"email" = :email,
		"roles" = :roles,
		"password_hash" = :password_hash,
		"date_updated" = :date_updated
	WHERE
		user_id = :user_id`

	s.log.Printf("%s: %s: %s", traceID, "user.Update",
		database.Log(q, usr),
	)

	if _, err := s.db.NamedExecContext(ctx, q, usr); err != nil {
		return errors.Wrapf(err, "updating user %s", usr.ID)
	}

	return nil
}

// Delete removes a user from the database.
func (s Store) Delete(ctx context.Context, traceID string, claims auth.Claims, userID string) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.delete")
	defer span.End()

	if err := validate.CheckID(userID); err != nil {
		return database.ErrInvalidID
	}

	// If you are not an admin and looking to delete someone other than yourself.
	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != userID {
		return database.ErrForbidden
	}

	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: userID,
	}

	const q = `
	DELETE FROM
		users
	WHERE
		user_id = :user_id`

	s.log.Printf("%s: %s: %s", traceID, "user.Delete",
		database.Log(q, data),
	)

	if _, err := s.db.NamedExecContext(ctx, q, data); err != nil {
		return errors.Wrapf(err, "deleting user %s", data.UserID)
	}

	return nil
}

// Query retrieves a list of existing users from the database.
func (s Store) Query(ctx context.Context, traceID string, pageNumber int, rowsPerPage int) ([]User, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.query")
	defer span.End()

	data := struct {
		Offset      int `db:"offset"`
		RowsPerPage int `db:"rows_per_page"`
	}{
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	const q = `
	SELECT
		*
	FROM
		users
	ORDER BY
		user_id
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`

	s.log.Printf("%s: %s: %s", traceID, "user.Query",
		database.Log(q, data),
	)

	var users []User
	if err := database.NamedQuerySlice(ctx, s.db, q, data, &users); err != nil {
		if err == database.ErrNotFound {
			return nil, database.ErrNotFound
		}
		return nil, errors.Wrap(err, "selecting users")
	}

	return users, nil
}

// QueryByID gets the specified user from the database.
func (s Store) QueryByID(ctx context.Context, traceID string, claims auth.Claims, userID string) (User, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.querybyid")
	defer span.End()

	if err := validate.CheckID(userID); err != nil {
		return User{}, database.ErrInvalidID
	}

	// If you are not an admin and looking to retrieve someone other than yourself.
	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != userID {
		return User{}, database.ErrForbidden
	}

	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: userID,
	}

	const q = `
	SELECT
		*
	FROM
		users
	WHERE 
		user_id = :user_id`

	s.log.Printf("%s: %s: %s", traceID, "user.QueryByID",
		database.Log(q, data),
	)

	var usr User
	if err := database.NamedQueryStruct(ctx, s.db, q, data, &usr); err != nil {
		if err == database.ErrNotFound {
			return User{}, database.ErrNotFound
		}
		return User{}, errors.Wrapf(err, "selecting user %q", data.UserID)
	}

	return usr, nil
}

// QueryByEmail gets the specified user from the database by email.
func (s Store) QueryByEmail(ctx context.Context, traceID string, claims auth.Claims, email string) (User, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.querybyemail")
	defer span.End()

	// Add Email Validate function in validate
	// if err := validate.Email(email); err != nil {
	// 	return User{}, ErrInvalidEmail
	// }

	data := struct {
		Email string `db:"email"`
	}{
		Email: email,
	}

	const q = `
	SELECT
		*
	FROM
		users
	WHERE
		email = :email`

	s.log.Printf("%s: %s: %s", traceID, "user.QueryByEmail",
		database.Log(q, data),
	)

	var usr User
	if err := database.NamedQueryStruct(ctx, s.db, q, data, &usr); err != nil {
		if err == database.ErrNotFound {
			return User{}, database.ErrNotFound
		}
		return User{}, errors.Wrapf(err, "selecting user %q", email)
	}

	// If you are not an admin and looking to retrieve someone other than yourself.
	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != usr.ID {
		return User{}, database.ErrForbidden
	}

	return usr, nil
}

// Authenticate finds a user by their email and verifies their password. On
// success it returns a Claims User representing this user. The claims can be
// used to generate a token for future authentication.
func (s Store) Authenticate(ctx context.Context, traceID string, now time.Time, email, password string) (auth.Claims, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.authenticate")
	defer span.End()

	data := struct {
		Email string `db:"email"`
	}{
		Email: email,
	}

	const q = `
	SELECT
		*
	FROM
		users
	WHERE
		email = :email`

	s.log.Printf("%s: %s: %s", traceID, "user.Authenticate",
		database.Log(q, data),
	)

	var usr User
	if err := database.NamedQueryStruct(ctx, s.db, q, data, &usr); err != nil {
		if err == database.ErrNotFound {
			return auth.Claims{}, database.ErrNotFound
		}
		return auth.Claims{}, errors.Wrapf(err, "selecting user %q", email)
	}

	// Compare the provided password with the saved hash. Use the bcrypt
	// comparison function so it is cryptographically secure.
	if err := bcrypt.CompareHashAndPassword(usr.PasswordHash, []byte(password)); err != nil {
		return auth.Claims{}, database.ErrAuthenticationFailure
	}

	// If we are this far the request is valid. Create some claims for the user
	// and generate their token.
	claims := auth.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    "service project",
			Subject:   usr.ID,
			ExpiresAt: jwt.At(now.Add(time.Hour)),
			IssuedAt:  jwt.At(now),
		},
		Roles: usr.Roles,
	}

	return claims, nil
}
