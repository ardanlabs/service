package user

import (
	"fmt"
	"net/mail"
	"time"

	"github.com/ardanlabs/service/business/sys/validate"
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
	Name            string
	Email           mail.Address
	Roles           []Role
	Department      string
	Password        string
	PasswordConfirm string
}

// Validate checks the data in the model is considered clean.
func (nu NewUser) Validate() error {
	if err := validate.Check(nu); err != nil {
		return err
	}
	return nil
}

// UpdateUser contains information needed to update a user.
type UpdateUser struct {
	Name            *string
	Email           *mail.Address
	Roles           []Role
	Department      *string
	Password        *string
	PasswordConfirm *string
	Enabled         *bool
}

// Validate checks the data in the model is considered clean.
func (uu UpdateUser) Validate() error {
	if err := validate.Check(uu); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}
