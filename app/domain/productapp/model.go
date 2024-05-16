package productapp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/app/sdk/mid"
	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/ardanlabs/service/foundation/validate"
)

// QueryParams represents the set of possible query strings.
type QueryParams struct {
	Page     string
	Rows     string
	OrderBy  string
	ID       string
	Name     string
	Cost     string
	Quantity string
}

// =============================================================================

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

// Encode implments the encoder interface.
func (app Product) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func toAppProduct(prd productbus.Product) Product {
	return Product{
		ID:          prd.ID.String(),
		UserID:      prd.UserID.String(),
		Name:        prd.Name.String(),
		Cost:        prd.Cost,
		Quantity:    prd.Quantity,
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

// Decode implments the decoder interface.
func (app *NewProduct) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app NewProduct) Validate() error {
	if err := validate.Check(app); err != nil {
		return errs.Newf(errs.FailedPrecondition, "validate: %s", err)
	}

	return nil
}

func toBusNewProduct(ctx context.Context, app NewProduct) (productbus.NewProduct, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return productbus.NewProduct{}, fmt.Errorf("getuserid: %w", err)
	}

	name, err := productbus.Names.Parse(app.Name)
	if err != nil {
		return productbus.NewProduct{}, fmt.Errorf("parse name: %w", err)
	}

	bus := productbus.NewProduct{
		UserID:   userID,
		Name:     name,
		Cost:     app.Cost,
		Quantity: app.Quantity,
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

// Decode implments the decoder interface.
func (app *UpdateProduct) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateProduct) Validate() error {
	if err := validate.Check(app); err != nil {
		return errs.Newf(errs.FailedPrecondition, "validate: %s", err)
	}

	return nil
}

func toBusUpdateProduct(app UpdateProduct) (productbus.UpdateProduct, error) {
	var name *productbus.Name
	if app.Name != nil {
		nm, err := productbus.Names.Parse(*app.Name)
		if err != nil {
			return productbus.UpdateProduct{}, fmt.Errorf("parse: %w", err)
		}
		name = &nm
	}

	bus := productbus.UpdateProduct{
		Name:     name,
		Cost:     app.Cost,
		Quantity: app.Quantity,
	}

	return bus, nil
}
