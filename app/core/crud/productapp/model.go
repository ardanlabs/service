package productapp

import (
	"context"
	"fmt"
	"time"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/business/core/crud/productbus"
	"github.com/ardanlabs/service/foundation/validate"
)

// QueryParams represents the set of possible query strings.
type QueryParams struct {
	Page     int    `query:"page"`
	Rows     int    `query:"rows"`
	OrderBy  string `query:"orderBy"`
	ID       string `query:"product_id"`
	Name     string `query:"name"`
	Cost     string `query:"cost"`
	Quantity string `query:"quantity"`
}

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

func toAppProduct(prd productbus.Product) Product {
	return Product{
		ID:          prd.ID.String(),
		UserID:      prd.UserID.String(),
		Name:        prd.Name,
		Cost:        prd.Cost,
		Quantity:    prd.Quantity,
		DateCreated: prd.DateCreated.Format(time.RFC3339),
		DateUpdated: prd.DateUpdated.Format(time.RFC3339),
	}
}

func toAppProducts(prds []productbus.Product) []Product {
	items := make([]Product, len(prds))
	for i, prd := range prds {
		items[i] = toAppProduct(prd)
	}

	return items
}

// NewProduct defines the data needed to add a new product.
type NewProduct struct {
	Name     string  `json:"name" validate:"required"`
	Cost     float64 `json:"cost" validate:"required,gte=0"`
	Quantity int     `json:"quantity" validate:"required,gte=1"`
}

func toBusNewProduct(ctx context.Context, app NewProduct) (productbus.NewProduct, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return productbus.NewProduct{}, fmt.Errorf("getuserid: %w", err)
	}

	prd := productbus.NewProduct{
		UserID:   userID,
		Name:     app.Name,
		Cost:     app.Cost,
		Quantity: app.Quantity,
	}

	return prd, nil
}

// Validate checks the data in the model is considered clean.
func (app NewProduct) Validate() error {
	if err := validate.Check(app); err != nil {
		return errs.Newf(errs.FailedPrecondition, "validate: %s", err)
	}

	return nil
}

// UpdateProduct defines the data needed to update a product.
type UpdateProduct struct {
	Name     *string  `json:"name"`
	Cost     *float64 `json:"cost" validate:"omitempty,gte=0"`
	Quantity *int     `json:"quantity" validate:"omitempty,gte=1"`
}

func toBusUpdateProduct(app UpdateProduct) productbus.UpdateProduct {
	core := productbus.UpdateProduct{
		Name:     app.Name,
		Cost:     app.Cost,
		Quantity: app.Quantity,
	}

	return core
}

// Validate checks the data in the model is considered clean.
func (app UpdateProduct) Validate() error {
	if err := validate.Check(app); err != nil {
		return errs.Newf(errs.FailedPrecondition, "validate: %s", err)
	}

	return nil
}
