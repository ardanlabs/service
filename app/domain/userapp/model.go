package userapp

import (
	"fmt"
	"net/mail"
	"time"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/foundation/validate"
)

// QueryParams represents the set of possible query strings.
type QueryParams struct {
	Page             int    `query:"page"`
	Rows             int    `query:"rows"`
	OrderBy          string `query:"orderBy"`
	ID               string `query:"user_id"`
	Name             string `query:"name"`
	Email            string `query:"email"`
	StartCreatedDate string `query:"start_created_date"`
	EndCreatedDate   string `query:"end_created_date"`
}

// User represents information about an individual user.
type User struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Email        string   `json:"email"`
	Roles        []string `json:"roles"`
	PasswordHash []byte   `json:"-"`
	Department   string   `json:"department"`
	Enabled      bool     `json:"enabled"`
	DateCreated  string   `json:"dateCreated"`
	DateUpdated  string   `json:"dateUpdated"`
}

func toAppUser(usr userbus.User) User {
	roles := make([]string, len(usr.Roles))
	for i, role := range usr.Roles {
		roles[i] = role.Name()
	}

	return User{
		ID:           usr.ID.String(),
		Name:         usr.Name,
		Email:        usr.Email.Address,
		Roles:        roles,
		PasswordHash: usr.PasswordHash,
		Department:   usr.Department,
		Enabled:      usr.Enabled,
		DateCreated:  usr.DateCreated.Format(time.RFC3339),
		DateUpdated:  usr.DateUpdated.Format(time.RFC3339),
	}
}

func toAppUsers(users []userbus.User) []User {
	app := make([]User, len(users))
	for i, usr := range users {
		app[i] = toAppUser(usr)
	}

	return app
}

// NewUser defines the data needed to add a new user.
type NewUser struct {
	Name            string   `json:"name" validate:"required"`
	Email           string   `json:"email" validate:"required,email"`
	Roles           []string `json:"roles" validate:"required"`
	Department      string   `json:"department"`
	Password        string   `json:"password" validate:"required"`
	PasswordConfirm string   `json:"passwordConfirm" validate:"eqfield=Password"`
}

func toBusNewUser(app NewUser) (userbus.NewUser, error) {
	roles := make([]userbus.Role, len(app.Roles))
	for i, roleStr := range app.Roles {
		role, err := userbus.ParseRole(roleStr)
		if err != nil {
			return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
		}
		roles[i] = role
	}

	addr, err := mail.ParseAddress(app.Email)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	bus := userbus.NewUser{
		Name:       app.Name,
		Email:      *addr,
		Roles:      roles,
		Department: app.Department,
		Password:   app.Password,
	}

	return bus, nil
}

// Validate checks the data in the model is considered clean.
func (app NewUser) Validate() error {
	if err := validate.Check(app); err != nil {
		return errs.Newf(errs.FailedPrecondition, "validate: %s", err)
	}

	return nil
}

// UpdateUserRole defines the data needed to update a user role.
type UpdateUserRole struct {
	Roles []string `json:"roles"`
}

func toBusUpdateUserRole(app UpdateUserRole) (userbus.UpdateUser, error) {
	var roles []userbus.Role
	if app.Roles != nil {
		roles = make([]userbus.Role, len(app.Roles))
		for i, roleStr := range app.Roles {
			role, err := userbus.ParseRole(roleStr)
			if err != nil {
				return userbus.UpdateUser{}, fmt.Errorf("parse: %w", err)
			}
			roles[i] = role
		}
	}

	bus := userbus.UpdateUser{
		Roles: roles,
	}

	return bus, nil
}

// UpdateUser defines the data needed to update a user.
type UpdateUser struct {
	Name            *string `json:"name"`
	Email           *string `json:"email" validate:"omitempty,email"`
	Department      *string `json:"department"`
	Password        *string `json:"password"`
	PasswordConfirm *string `json:"passwordConfirm" validate:"omitempty,eqfield=Password"`
	Enabled         *bool   `json:"enabled"`
}

func toBusUpdateUser(app UpdateUser) (userbus.UpdateUser, error) {
	var addr *mail.Address
	if app.Email != nil {
		var err error
		addr, err = mail.ParseAddress(*app.Email)
		if err != nil {
			return userbus.UpdateUser{}, fmt.Errorf("parse: %w", err)
		}
	}

	bus := userbus.UpdateUser{
		Name:       app.Name,
		Email:      addr,
		Department: app.Department,
		Password:   app.Password,
		Enabled:    app.Enabled,
	}

	return bus, nil
}

// Validate checks the data in the model is considered clean.
func (app UpdateUser) Validate() error {
	if err := validate.Check(app); err != nil {
		return errs.Newf(errs.FailedPrecondition, "validate: %s", err)
	}

	return nil
}
