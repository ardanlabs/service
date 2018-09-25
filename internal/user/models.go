package user

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// User represents someone with access to our system.
type User struct {
	ID    bson.ObjectId `bson:"_id" json:"id"`
	Name  string        `bson:"name" json:"name"`
	Email string        `bson:"email" json:"email"` // TODO(jlw) enforce uniqueness
	Roles []string      `bson:"roles" json:"roles"`

	PasswordHash []byte `bson:"password_hash" json:"-"`

	DateModified time.Time `bson:"date_modified" json:"date_modified"`
	DateCreated  time.Time `bson:"date_created,omitempty" json:"date_created"`
}

// NewUser contains information needed to create a new User.
type NewUser struct {
	Name            string   `json:"name" validate:"required"`
	Email           string   `json:"email" validate:"required"` // TODO(jlw) enforce uniqueness.
	Roles           []string `json:"roles" validate:"required"` // TODO(jlw) Ensure only includes valid roles.
	Password        string   `json:"password" validate:"required"`
	PasswordConfirm string   `json:"password_confirm" validate:"eqfield=Password"`
}

// UpdateUser defines what information may be provided to modify an existing
// User. All fields are optional so clients can send just the fields they want
// changed. It uses pointer fields so we can differentiate between a field that
// was not provided and a field that was provided as explicitly blank. Normally
// we do not want to use pointers to basic types but we make exceptions around
// marshalling/unmarshalling.
type UpdateUser struct {
	Name            *string  `json:"name"`
	Email           *string  `json:"email"` // TODO(jlw) enforce uniqueness.
	Roles           []string `json:"roles"` // TODO(jlw) Ensure only includes valid roles.
	Password        *string  `json:"password"`
	PasswordConfirm *string  `json:"password_confirm" validate:"omitempty,eqfield=Password"`
}

// Token is the payload we deliver to users when they authenticate.
type Token struct {
	Token string `json:"token"`
}
