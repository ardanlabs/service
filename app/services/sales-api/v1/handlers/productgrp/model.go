package productgrp

import (
	"context"
	"time"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/web/v1/mid"
	"github.com/ardanlabs/service/foundation/validate"
	"github.com/google/uuid"
)

// AppProduct represents an individual product.
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

// AppProductDetails represents an individual product.
type AppProductDetails struct {
	ID          string  `json:"id"`
	UserID      string  `json:"userID"`
	Name        string  `json:"name"`
	Cost        float64 `json:"cost"`
	Quantity    int     `json:"quantity"`
	UserName    string  `json:"userName"`
	DateCreated string  `json:"dateCreated"`
	DateUpdated string  `json:"dateUpdated"`
}

func toAppProductDetails(prd product.Product, usr user.User) AppProductDetails {
	return AppProductDetails{
		ID:          prd.ID.String(),
		Name:        prd.Name,
		Cost:        prd.Cost,
		Quantity:    prd.Quantity,
		UserID:      prd.UserID.String(),
		UserName:    usr.Name,
		DateCreated: prd.DateCreated.Format(time.RFC3339),
		DateUpdated: prd.DateUpdated.Format(time.RFC3339),
	}
}

func toAppProductsDetails(prds []product.Product, usrs map[uuid.UUID]user.User) []AppProductDetails {
	items := make([]AppProductDetails, len(prds))
	for i, prd := range prds {
		items[i] = toAppProductDetails(prd, usrs[prd.UserID])
	}

	return items
}

// AppNewProduct is what we require from clients when adding a Product.
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

// AppUpdateProduct contains information needed to update a product.
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
