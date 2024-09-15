package userbus

import (
	"net/mail"
	"time"

	"github.com/ardanlabs/service/business/types/name"
	"github.com/ardanlabs/service/business/types/role"
	"github.com/google/uuid"
)

// User represents information about an individual user.
type User struct {
	ID           uuid.UUID
	Name         name.Name
	Email        mail.Address
	Roles        []role.Role
	PasswordHash []byte
	Department   name.Null
	Enabled      bool
	DateCreated  time.Time
	DateUpdated  time.Time
}

// NewUser contains information needed to create a new user.
type NewUser struct {
	Name       name.Name
	Email      mail.Address
	Roles      []role.Role
	Department name.Null
	Password   string
}

// UpdateUser contains information needed to update a user.
type UpdateUser struct {
	Name       *name.Name
	Email      *mail.Address
	Roles      []role.Role
	Department *name.Null
	Password   *string
	Enabled    *bool
}
