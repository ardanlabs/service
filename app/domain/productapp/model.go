package productapp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/app/sdk/mid"
	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/ardanlabs/service/business/types/money"
	"github.com/ardanlabs/service/business/types/name"
	"github.com/ardanlabs/service/business/types/quantity"
)

// Product represents information about an individual product.
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

func toAppProducts(prds []productbus.Product) []Product {
	app := make([]Product, len(prds))
	for i, prd := range prds {
		app[i] = toAppProduct(prd)
	}

	return app
}

// =============================================================================

// NewProduct defines the data needed to add a new product.
type NewProduct struct {
	Name     string  `json:"name" validate:"required"`
	Cost     float64 `json:"cost" validate:"required,gte=0"`
	Quantity int     `json:"quantity" validate:"required,gte=1"`
}

// Decode implements the decoder interface.
func (app *NewProduct) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app NewProduct) Validate() error {
	if err := errs.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

func toBusNewProduct(ctx context.Context, app NewProduct) (productbus.NewProduct, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return productbus.NewProduct{}, fmt.Errorf("getuserid: %w", err)
	}

	name, err := name.Parse(app.Name)
	if err != nil {
		return productbus.NewProduct{}, fmt.Errorf("parse name: %w", err)
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
		UserID:   userID,
		Name:     name,
		Cost:     cost,
		Quantity: quantity,
	}

	return bus, nil
}

// =============================================================================

// UpdateProduct defines the data needed to update a product.
type UpdateProduct struct {
	Name     *string  `json:"name"`
	Cost     *float64 `json:"cost" validate:"omitempty,gte=0"`
	Quantity *int     `json:"quantity" validate:"omitempty,gte=1"`
}

// Decode implements the decoder interface.
func (app *UpdateProduct) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateProduct) Validate() error {
	if err := errs.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

func toBusUpdateProduct(app UpdateProduct) (productbus.UpdateProduct, error) {
	var nme *name.Name
	if app.Name != nil {
		nm, err := name.Parse(*app.Name)
		if err != nil {
			return productbus.UpdateProduct{}, fmt.Errorf("parse: %w", err)
		}
		nme = &nm
	}

	var cost *money.Money
	if app.Cost != nil {
		cst, err := money.Parse(*app.Cost)
		if err != nil {
			return productbus.UpdateProduct{}, fmt.Errorf("parse: %w", err)
		}
		cost = &cst
	}

	var qnt *quantity.Quantity
	if app.Cost != nil {
		qn, err := quantity.Parse(*app.Quantity)
		if err != nil {
			return productbus.UpdateProduct{}, fmt.Errorf("parse: %w", err)
		}
		qnt = &qn
	}

	bus := productbus.UpdateProduct{
		Name:     nme,
		Cost:     cost,
		Quantity: qnt,
	}

	return bus, nil
}
