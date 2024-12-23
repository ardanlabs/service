package tranapp

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"time"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/business/types/money"
	"github.com/ardanlabs/service/business/types/name"
	"github.com/ardanlabs/service/business/types/quantity"
	"github.com/ardanlabs/service/business/types/role"
)

// Product represents an individual product.
type Product struct {
	ID          string  `json:"id"`
	UserID      string  `json:"userID"`
	Name        string  `json:"name"`
	Cost        float64 `json:"cost"`
	Quantity    int     `json:"quantity"`
	DateCreated string  `json:"dateCreated"`
	DateUpdated string  `json:"dateUpdated"`
}

// Encode implements the encoder interface.
func (app Product) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func toAppProduct(prd productbus.Product) Product {
	return Product{
		ID:          prd.ID.String(),
		UserID:      prd.UserID.String(),
		Name:        prd.Name.String(),
		Cost:        prd.Cost.Value(),
		Quantity:    prd.Quantity.Value(),
		DateCreated: prd.DateCreated.Format(time.RFC3339),
		DateUpdated: prd.DateUpdated.Format(time.RFC3339),
	}
}

// =============================================================================

// NewTran represents an example of cross domain transaction at the
// application layer.
type NewTran struct {
	Product NewProduct `json:"product"`
	User    NewUser    `json:"user"`
}

// Validate checks the data in the model is considered clean.
func (app NewTran) Validate() error {
	if err := errs.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

// Decode implements the decoder interface.
func (app *NewTran) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// =============================================================================

// NewUser contains information needed to create a new user.
type NewUser struct {
	Name            string   `json:"name" validate:"required"`
	Email           string   `json:"email" validate:"required,email"`
	Roles           []string `json:"roles" validate:"required"`
	Department      string   `json:"department"`
	Password        string   `json:"password" validate:"required"`
	PasswordConfirm string   `json:"passwordConfirm" validate:"eqfield=Password"`
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

// NewProduct is what we require from clients when adding a Product.
type NewProduct struct {
	Name     string  `json:"name" validate:"required"`
	Cost     float64 `json:"cost" validate:"required,gte=0"`
	Quantity int     `json:"quantity" validate:"required,gte=1"`
}

// Validate checks the data in the model is considered clean.
func (app NewProduct) Validate() error {
	if err := errs.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

func toBusNewProduct(app NewProduct) (productbus.NewProduct, error) {
	name, err := name.Parse(app.Name)
	if err != nil {
		return productbus.NewProduct{}, fmt.Errorf("parse: %w", err)
	}

	cost, err := money.Parse(app.Cost)
	if err != nil {
		return productbus.NewProduct{}, fmt.Errorf("parse cost: %w", err)
	}

	quantity, err := quantity.Parse(app.Quantity)
	if err != nil {
		return productbus.NewProduct{}, fmt.Errorf("parse quantity: %w", err)
	}

	bus := productbus.NewProduct{
		Name:     name,
		Cost:     cost,
		Quantity: quantity,
	}

	return bus, nil
}
