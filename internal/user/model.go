package user

import "time"

// CreateUser contains information needed to create or update a user.
type CreateUser struct {
	UserType     int        `json:"type" validate:"required"`
	FirstName    string     `json:"first_name" validate:"required"`
	LastName     string     `json:"last_name" validate:"required"`
	Email        string     `json:"email" validate:"required"`
	Company      string     `json:"company" validate:"required"`
	DateModified *time.Time `json:"-"`
}

// User contains information about a user.
type User struct {
	UserID       string     `bson:"user_id,omitempty" json:"user_id,omitempty"`
	UserType     int        `bson:"type" json:"type"`
	FirstName    string     `bson:"first_name" json:"first_name"`
	LastName     string     `bson:"last_name" json:"last_name"`
	Email        string     `bson:"email" json:"email"`
	Company      string     `bson:"company" json:"company"`
	DateModified *time.Time `bson:"date_modified" json:"date_modified"`
	DateCreated  *time.Time `bson:"date_created,omitempty" json:"date_created"`
}
