package userbus

import (
	"net/mail"
	"time"

	"github.com/google/uuid"
)

// User represents information about an individual user.
type User struct {
	ID           uuid.UUID
	Name         string
	Email        mail.Address
	Roles        []Role
	PasswordHash []byte
	Department   string
	Enabled      bool
	DateCreated  time.Time
	DateUpdated  time.Time
}

// NewUser contains information needed to create a new user.
type NewUser struct {
	Name       string
	Email      mail.Address
	Roles      []Role
	Department string
	Password   string
}

// UpdateUser contains information needed to update a user.
type UpdateUser struct {
	Name       *string
	Email      *mail.Address
	Roles      []Role
	Department *string
	Password   *string
	Enabled    *bool
}
