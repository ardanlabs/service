package vproductgrp

import (
	"time"

	"github.com/ardanlabs/service/business/core/views/vproduct"
)

// AppProduct represents information about an individual product with
// extended information.
type AppProduct struct {
	ID          string  `json:"id"`
	UserID      string  `json:"userID"`
	Name        string  `json:"name"`
	Cost        float64 `json:"cost"`
	Quantity    int     `json:"quantity"`
	DateCreated string  `json:"dateCreated"`
	DateUpdated string  `json:"dateUpdated"`
	UserName    string  `json:"userName"`
}

func toAppProduct(prd vproduct.Product) AppProduct {
	return AppProduct{
		ID:          prd.ID.String(),
		UserID:      prd.UserID.String(),
		Name:        prd.Name,
		Cost:        prd.Cost,
		Quantity:    prd.Quantity,
		DateCreated: prd.DateCreated.Format(time.RFC3339),
		DateUpdated: prd.DateUpdated.Format(time.RFC3339),
		UserName:    prd.UserName,
	}
}

func toAppProducts(prds []vproduct.Product) []AppProduct {
	items := make([]AppProduct, len(prds))
	for i, prd := range prds {
		items[i] = toAppProduct(prd)
	}

	return items
}
