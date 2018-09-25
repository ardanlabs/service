package user

import (
	"context"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/bcrypt"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const usersCollection = "users"

var (
	// ErrNotFound abstracts the mgo not found error.
	ErrNotFound = errors.New("Entity not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrAuthenticationFailure occurs when a user attempts to authenticate but
	// anything goes wrong.
	ErrAuthenticationFailure = errors.New("Authentication failed")
)

// List retrieves a list of existing users from the database.
func List(ctx context.Context, dbConn *db.DB) ([]User, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.List")
	defer span.End()

	u := []User{}

	f := func(collection *mgo.Collection) error {
		return collection.Find(nil).All(&u)
	}
	if err := dbConn.Execute(ctx, usersCollection, f); err != nil {
		return nil, errors.Wrap(err, "db.users.find()")
	}

	return u, nil
}

// Retrieve gets the specified user from the database.
func Retrieve(ctx context.Context, dbConn *db.DB, id string) (*User, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.Retrieve")
	defer span.End()

	if !bson.IsObjectIdHex(id) {
		return nil, ErrInvalidID
	}

	q := bson.M{"_id": bson.ObjectIdHex(id)}

	var u *User
	f := func(collection *mgo.Collection) error {
		return collection.Find(q).One(&u)
	}
	if err := dbConn.Execute(ctx, usersCollection, f); err != nil {
		if err == mgo.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, fmt.Sprintf("db.users.find(%s)", db.Query(q)))
	}

	return u, nil
}

// Create inserts a new user into the database.
func Create(ctx context.Context, dbConn *db.DB, nu *NewUser, now time.Time) (*User, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.Create")
	defer span.End()

	// Mongo truncates times to milliseconds when storing. We and do the same
	// here so the value we return is consistent with what we store.
	now = now.Truncate(time.Millisecond)

	pw, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, "generating password hash")
	}

	u := User{
		ID:           bson.NewObjectId(),
		Name:         nu.Name,
		Email:        nu.Email,
		PasswordHash: pw,
		Roles:        nu.Roles,
		DateCreated:  now,
		DateModified: now,
	}

	f := func(collection *mgo.Collection) error {
		return collection.Insert(&u)
	}
	if err := dbConn.Execute(ctx, usersCollection, f); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("db.users.insert(%s)", db.Query(&u)))
	}

	return &u, nil
}

// Update replaces a user document in the database.
func Update(ctx context.Context, dbConn *db.DB, id string, upd *UpdateUser, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.user.Update")
	defer span.End()

	if !bson.IsObjectIdHex(id) {
		return ErrInvalidID
	}

	fields := make(bson.M)

	if upd.Name != nil {
		fields["name"] = *upd.Name
	}
	if upd.Email != nil {
		fields["email"] = *upd.Email
	}
	if upd.Roles != nil {
		fields["roles"] = upd.Roles
	}
	if upd.Password != nil {
		pw, err := bcrypt.GenerateFromPassword([]byte(*upd.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.Wrap(err, "generating password hash")
		}
		fields["password_hash"] = pw
	}

	// If there's nothing to update we can quit early.
	if len(fields) == 0 {
		return nil
	}

	fields["date_modified"] = now

	m := bson.M{"$set": fields}
	q := bson.M{"_id": bson.ObjectIdHex(id)}

	f := func(collection *mgo.Collection) error {
		return collection.Update(q, m)
	}
	if err := dbConn.Execute(ctx, usersCollection, f); err != nil {
		if err == mgo.ErrNotFound {
			return ErrNotFound
		}
		return errors.Wrap(err, fmt.Sprintf("db.customers.update(%s, %s)", db.Query(q), db.Query(m)))
	}

	return nil
}

// Delete removes a user from the database.
func Delete(ctx context.Context, dbConn *db.DB, id string) error {
	ctx, span := trace.StartSpan(ctx, "internal.user.Update")
	defer span.End()

	if !bson.IsObjectIdHex(id) {
		return ErrInvalidID
	}

	q := bson.M{"_id": bson.ObjectIdHex(id)}

	f := func(collection *mgo.Collection) error {
		return collection.Remove(q)
	}
	if err := dbConn.Execute(ctx, usersCollection, f); err != nil {
		if err == mgo.ErrNotFound {
			return ErrNotFound
		}
		return errors.Wrap(err, fmt.Sprintf("db.users.remove(%s)", db.Query(q)))
	}

	return nil
}

// Authenticate finds a user by their email and verifies their password. On
// success it returns a Token that can be used to authenticate in the future.
//
// The key, keyID, and alg are required for generating the token.
func Authenticate(ctx context.Context, dbConn *db.DB, now time.Time, key *rsa.PrivateKey, keyID, alg, email, password string) (Token, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.Authenticate")
	defer span.End()

	q := bson.M{"email": email}

	var u *User
	f := func(collection *mgo.Collection) error {
		return collection.Find(q).One(&u)
	}
	if err := dbConn.Execute(ctx, usersCollection, f); err != nil {

		// Normally we would return ErrNotFound in this scenario but we do not want
		// to leak to an unauthenticated user which emails are in the system.
		if err == mgo.ErrNotFound {
			return Token{}, ErrAuthenticationFailure
		}
		return Token{}, errors.Wrap(err, fmt.Sprintf("db.users.find(%s)", db.Query(q)))
	}

	// Compare the provided password with the saved hash. Use the bcrypt
	// comparison function so it is cryptographically secure.
	if err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password)); err != nil {
		return Token{}, ErrAuthenticationFailure
	}

	// If we are this far the request is valid. Create some claims for the user
	// and generate their token.
	claims := auth.NewClaims(u.ID.Hex(), u.Roles, now, time.Hour)

	tkn, err := auth.GenerateToken(key, keyID, alg, claims)
	if err != nil {
		return Token{}, errors.Wrap(err, "generating token")
	}

	return Token{Token: tkn}, nil
}
