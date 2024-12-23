package userapp

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"time"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/business/types/name"
	"github.com/ardanlabs/service/business/types/role"
)

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

// Encode implements the encoder interface.
func (app User) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func toAppUser(bus userbus.User) User {
	return User{
		ID:           bus.ID.String(),
		Name:         bus.Name.String(),
		Email:        bus.Email.Address,
		Roles:        role.ParseToString(bus.Roles),
		PasswordHash: bus.PasswordHash,
		Department:   bus.Department.String(),
		Enabled:      bus.Enabled,
		DateCreated:  bus.DateCreated.Format(time.RFC3339),
		DateUpdated:  bus.DateUpdated.Format(time.RFC3339),
	}
}

func toAppUsers(users []userbus.User) []User {
	app := make([]User, len(users))
	for i, usr := range users {
		app[i] = toAppUser(usr)
	}

	return app
}

// =============================================================================

// NewUser defines the data needed to add a new user.
type NewUser struct {
	Name            string   `json:"name" validate:"required"`
	Email           string   `json:"email" validate:"required,email"`
	Roles           []string `json:"roles" validate:"required"`
	Department      string   `json:"department"`
	Password        string   `json:"password" validate:"required"`
	PasswordConfirm string   `json:"passwordConfirm" validate:"eqfield=Password"`
}

// Decode implements the decoder interface.
func (app *NewUser) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app NewUser) Validate() error {
	if err := errs.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

func toBusNewUser(app NewUser) (userbus.NewUser, error) {
	roles, err := role.ParseMany(app.Roles)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	addr, err := mail.ParseAddress(app.Email)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	nme, err := name.Parse(app.Name)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	department, err := name.ParseNull(app.Department)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	bus := userbus.NewUser{
		Name:       nme,
		Email:      *addr,
		Roles:      roles,
		Department: department,
		Password:   app.Password,
	}

	return bus, nil
}

// =============================================================================

// UpdateUserRole defines the data needed to update a user role.
type UpdateUserRole struct {
	Roles []string `json:"roles" validate:"required"`
}

// Decode implements the decoder interface.
func (app *UpdateUserRole) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateUserRole) Validate() error {
	if err := errs.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

func toBusUpdateUserRole(app UpdateUserRole) (userbus.UpdateUser, error) {
	var roles []role.Role
	if app.Roles != nil {
		var err error
		roles, err = role.ParseMany(app.Roles)
		if err != nil {
			return userbus.UpdateUser{}, fmt.Errorf("parse: %w", err)
		}
	}

	bus := userbus.UpdateUser{
		Roles: roles,
	}

	return bus, nil
}

// =============================================================================

// UpdateUser defines the data needed to update a user.
type UpdateUser struct {
	Name            *string `json:"name"`
	Email           *string `json:"email" validate:"omitempty,email"`
	Department      *string `json:"department"`
	Password        *string `json:"password"`
	PasswordConfirm *string `json:"passwordConfirm" validate:"omitempty,eqfield=Password"`
	Enabled         *bool   `json:"enabled"`
}

// Decode implements the decoder interface.
func (app *UpdateUser) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateUser) Validate() error {
	if err := errs.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
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

	var nme *name.Name
	if app.Name != nil {
		nm, err := name.Parse(*app.Name)
		if err != nil {
			return userbus.UpdateUser{}, fmt.Errorf("parse: %w", err)
		}
		nme = &nm
	}

	var department *name.Null
	if app.Department != nil {
		dep, err := name.ParseNull(*app.Department)
		if err != nil {
			return userbus.UpdateUser{}, fmt.Errorf("parse: %w", err)
		}
		department = &dep
	}

	bus := userbus.UpdateUser{
		Name:       nme,
		Email:      addr,
		Department: department,
		Password:   app.Password,
		Enabled:    app.Enabled,
	}

	return bus, nil
}
