package productgrp

import (
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/sys/validate"
	"github.com/google/uuid"
)

// AppProduct represents an individual product.
type AppProduct struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Cost        int    `json:"cost"`
	Quantity    int    `json:"quantity"`
	Sold        int    `json:"sold"`
	Revenue     int    `json:"revenue"`
	UserID      string `json:"userID"`
	DateCreated string `json:"dateCreated"`
	DateUpdated string `json:"dateUpdated"`
}

func toAppProduct(core product.Product) AppProduct {
	return AppProduct{
		ID:          core.ID.String(),
		Name:        core.Name,
		Cost:        core.Cost,
		Quantity:    core.Quantity,
		Sold:        core.Sold,
		Revenue:     core.Revenue,
		UserID:      core.UserID.String(),
		DateCreated: core.DateCreated.Format(time.RFC3339),
		DateUpdated: core.DateUpdated.Format(time.RFC3339),
	}
}

// =============================================================================

// AppNewProduct is what we require from clients when adding a Product.
type AppNewProduct struct {
	Name     string `json:"name" validate:"required"`
	Cost     int    `json:"cost" validate:"required,gte=0"`
	Quantity int    `json:"quantity" validate:"gte=1"`
	UserID   string `json:"userID" validate:"required"`
}

func toCoreNewProduct(app AppNewProduct) (product.NewProduct, error) {
	userID, err := uuid.Parse(app.UserID)
	if err != nil {
		return product.NewProduct{}, fmt.Errorf("parsing userid: %w", err)
	}

	core := product.NewProduct{
		Name:     app.Name,
		Cost:     app.Cost,
		Quantity: app.Quantity,
		UserID:   userID,
	}

	return core, nil
}

// Validate checks the data in the model is considered clean.
func (app AppNewProduct) Validate() error {
	if err := validate.Check(app); err != nil {
		return err
	}
	return nil
}

// =============================================================================

// AppUpdateProduct contains information needed to update a product.
type AppUpdateProduct struct {
	Name     *string `json:"name"`
	Cost     *int    `json:"cost" validate:"omitempty,gte=0"`
	Quantity *int    `json:"quantity" validate:"omitempty,gte=1"`
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
