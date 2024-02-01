package productgrp

import (
	"context"
	"time"

	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/web/v1/mid"
	"github.com/ardanlabs/service/foundation/validate"
)

// AppProduct represents information about an individual product.
type AppProduct struct {
	ID          string  `json:"id"`
	UserID      string  `json:"userID"`
	Name        string  `json:"name"`
	Cost        float64 `json:"cost"`
	Quantity    int     `json:"quantity"`
	DateCreated string  `json:"dateCreated"`
	DateUpdated string  `json:"dateUpdated"`
}

func toAppProduct(prd product.Product) AppProduct {
	return AppProduct{
		ID:          prd.ID.String(),
		UserID:      prd.UserID.String(),
		Name:        prd.Name,
		Cost:        prd.Cost,
		Quantity:    prd.Quantity,
		DateCreated: prd.DateCreated.Format(time.RFC3339),
		DateUpdated: prd.DateUpdated.Format(time.RFC3339),
	}
}

func toAppProducts(prds []product.Product) []AppProduct {
	items := make([]AppProduct, len(prds))
	for i, prd := range prds {
		items[i] = toAppProduct(prd)
	}

	return items
}

// AppNewProduct defines the data needed to add a new product.
type AppNewProduct struct {
	Name     string  `json:"name" validate:"required"`
	Cost     float64 `json:"cost" validate:"required,gte=0"`
	Quantity int     `json:"quantity" validate:"required,gte=1"`
}

func toCoreNewProduct(ctx context.Context, app AppNewProduct) product.NewProduct {
	prd := product.NewProduct{
		UserID:   mid.GetUserID(ctx),
		Name:     app.Name,
		Cost:     app.Cost,
		Quantity: app.Quantity,
	}

	return prd
}

// Validate checks the data in the model is considered clean.
func (app AppNewProduct) Validate() error {
	if err := validate.Check(app); err != nil {
		return err
	}

	return nil
}

// AppUpdateProduct defines the data needed to update a product.
type AppUpdateProduct struct {
	Name     *string  `json:"name"`
	Cost     *float64 `json:"cost" validate:"omitempty,gte=0"`
	Quantity *int     `json:"quantity" validate:"omitempty,gte=1"`
}

func toCoreUpdateProduct(app AppUpdateProduct) product.UpdateProduct {
	core := product.UpdateProduct{
		Name:     app.Name,
		Cost:     app.Cost,
		Quantity: app.Quantity,
	}

	return core
}

// Validate checks the data in the model is considered clean.
func (app AppUpdateProduct) Validate() error {
	if err := validate.Check(app); err != nil {
		return err
	}

	return nil
}
