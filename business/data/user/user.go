// Package user contains user related CRUD functionality.
package user

import (
	"context"
	"log"
	"time"

	"github.com/ardanlabs/service/business/auth"
	"github.com/ardanlabs/service/business/validate"
	"github.com/ardanlabs/service/foundation/database"
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrNotFound is used when a specific User is requested but does not exist.
	ErrNotFound = errors.New("not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrAuthenticationFailure occurs when a user attempts to authenticate but
	// anything goes wrong.
	ErrAuthenticationFailure = errors.New("authentication failed")

	// ErrForbidden occurs when a user tries to do something that is forbidden to them according to our access control policies.
	ErrForbidden = errors.New("attempted action is not allowed")
)

// User manages the set of API's for user access.
type User struct {
	log *log.Logger
	db  *sqlx.DB
}

// New constructs a User for api access.
func New(log *log.Logger, db *sqlx.DB) User {
	return User{
		log: log,
		db:  db,
	}
}

// Create inserts a new user into the database.
func (u User) Create(ctx context.Context, traceID string, nu NewUser, now time.Time) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.create")
	defer span.End()

	if err := validate.Check(nu); err != nil {
		return Info{}, errors.Wrap(err, "validating data")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)
	if err != nil {
		return Info{}, errors.Wrap(err, "generating password hash")
	}

	usr := Info{
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

	u.log.Printf("%s: %s: %s", traceID, "user.Create",
		database.Log(q, usr),
	)

	if _, err := u.db.NamedExecContext(ctx, q, usr); err != nil {
		return Info{}, errors.Wrap(err, "inserting user")
	}

	return usr, nil
}

// Update replaces a user document in the database.
func (u User) Update(ctx context.Context, traceID string, claims auth.Claims, userID string, uu UpdateUser, now time.Time) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.update")
	defer span.End()

	if err := validate.CheckID(userID); err != nil {
		return ErrInvalidID
	}
	if err := validate.Check(uu); err != nil {
		return errors.Wrap(err, "validating data")
	}

	usr, err := u.QueryByID(ctx, traceID, claims, userID)
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

	u.log.Printf("%s: %s: %s", traceID, "user.Update",
		database.Log(q, usr),
	)

	if _, err := u.db.NamedExecContext(ctx, q, usr); err != nil {
		return errors.Wrapf(err, "updating user %s", usr.ID)
	}

	return nil
}

// Delete removes a user from the database.
func (u User) Delete(ctx context.Context, traceID string, claims auth.Claims, userID string) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.delete")
	defer span.End()

	if err := validate.CheckID(userID); err != nil {
		return ErrInvalidID
	}

	// If you are not an admin and looking to delete someone other than yourself.
	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != userID {
		return ErrForbidden
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

	u.log.Printf("%s: %s: %s", traceID, "user.Delete",
		database.Log(q, data),
	)

	if _, err := u.db.NamedExecContext(ctx, q, data); err != nil {
		return errors.Wrapf(err, "deleting user %s", data.UserID)
	}

	return nil
}

// Query retrieves a list of existing users from the database.
func (u User) Query(ctx context.Context, traceID string, pageNumber int, rowsPerPage int) ([]Info, error) {
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

	u.log.Printf("%s: %s: %s", traceID, "user.Query",
		database.Log(q, data),
	)

	var users []Info
	if err := database.NamedQuerySlice(ctx, u.db, q, data, &users); err != nil {
		if err == database.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "selecting users")
	}

	return users, nil
}

// QueryByID gets the specified user from the database.
func (u User) QueryByID(ctx context.Context, traceID string, claims auth.Claims, userID string) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.querybyid")
	defer span.End()

	if err := validate.CheckID(userID); err != nil {
		return Info{}, ErrInvalidID
	}

	// If you are not an admin and looking to retrieve someone other than yourself.
	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != userID {
		return Info{}, ErrForbidden
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

	u.log.Printf("%s: %s: %s", traceID, "user.QueryByID",
		database.Log(q, data),
	)

	var usr Info
	if err := database.NamedQueryStruct(ctx, u.db, q, data, &usr); err != nil {
		if err == database.ErrNotFound {
			return Info{}, ErrNotFound
		}
		return Info{}, errors.Wrapf(err, "selecting user %q", data.UserID)
	}

	return usr, nil
}

// QueryByEmail gets the specified user from the database by email.
func (u User) QueryByEmail(ctx context.Context, traceID string, claims auth.Claims, email string) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.querybyemail")
	defer span.End()

	// Add Email Validate function in validate
	// if err := validate.Email(email); err != nil {
	// 	return Info{}, ErrInvalidEmail
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

	u.log.Printf("%s: %s: %s", traceID, "user.QueryByEmail",
		database.Log(q, data),
	)

	var usr Info
	if err := database.NamedQueryStruct(ctx, u.db, q, data, &usr); err != nil {
		if err == database.ErrNotFound {
			return Info{}, ErrNotFound
		}
		return Info{}, errors.Wrapf(err, "selecting user %q", email)
	}

	// If you are not an admin and looking to retrieve someone other than yourself.
	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != usr.ID {
		return Info{}, ErrForbidden
	}

	return usr, nil
}

// Authenticate finds a user by their email and verifies their password. On
// success it returns a Claims Info representing this user. The claims can be
// used to generate a token for future authentication.
func (u User) Authenticate(ctx context.Context, traceID string, now time.Time, email, password string) (auth.Claims, error) {
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

	u.log.Printf("%s: %s: %s", traceID, "user.Authenticate",
		database.Log(q, data),
	)

	var usr Info
	if err := database.NamedQueryStruct(ctx, u.db, q, data, &usr); err != nil {
		if err == database.ErrNotFound {
			return auth.Claims{}, ErrNotFound
		}
		return auth.Claims{}, errors.Wrapf(err, "selecting user %q", email)
	}

	// Compare the provided password with the saved hash. Use the bcrypt
	// comparison function so it is cryptographically secure.
	if err := bcrypt.CompareHashAndPassword(usr.PasswordHash, []byte(password)); err != nil {
		return auth.Claims{}, ErrAuthenticationFailure
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
