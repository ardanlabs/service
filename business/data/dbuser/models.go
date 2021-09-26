package dbuser

import (
	"time"

	"github.com/lib/pq"
)

// DBUser represent the structure we need for moving data
// between the app and the database.
type DBUser struct {
	ID           string         `db:"user_id"`
	Name         string         `db:"name"`
	Email        string         `db:"email"`
	Roles        pq.StringArray `db:"roles"`
	PasswordHash []byte         `db:"password_hash"`
	DateCreated  time.Time      `db:"date_created"`
	DateUpdated  time.Time      `db:"date_updated"`
}

// DBNewUser contains information needed to create a new User.
type DBNewUser struct {
	Name            string
	Email           string
	Roles           []string
	Password        string
	PasswordConfirm string
}

// DBUpdateUser defines what information may be provided to modify an existing
// User. All fields are optional so clients can send just the fields they want
// changed. It uses pointer fields so we can differentiate between a field that
// was not provided and a field that was provided as explicitly blank. Normally
// we do not want to use pointers to basic types but we make exceptions around
// marshalling/unmarshalling.
type DBUpdateUser struct {
	Name            *string
	Email           *string
	Roles           []string
	Password        *string
	PasswordConfirm *string
}
