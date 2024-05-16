package vproductapp

import (
	"encoding/json"
	"time"

	"github.com/ardanlabs/service/business/domain/vproductbus"
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
	UserName string
}

// =============================================================================

// Product represents information about an individual product with
// extended information.
type Product struct {
	ID          string  `json:"id"`
	UserID      string  `json:"userID"`
	Name        string  `json:"name"`
	Cost        float64 `json:"cost"`
	Quantity    int     `json:"quantity"`
	DateCreated string  `json:"dateCreated"`
	DateUpdated string  `json:"dateUpdated"`
	UserName    string  `json:"userName"`
}

// Encode implments the encoder interface.
func (app Product) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func toAppProduct(prd vproductbus.Product) Product {
	return Product{
		ID:          prd.ID.String(),
		UserID:      prd.UserID.String(),
		Name:        prd.Name.String(),
		Cost:        prd.Cost,
		Quantity:    prd.Quantity,
		DateCreated: prd.DateCreated.Format(time.RFC3339),
		DateUpdated: prd.DateUpdated.Format(time.RFC3339),
		UserName:    prd.UserName.String(),
	}
}

func toAppProducts(prds []vproductbus.Product) []Product {
	app := make([]Product, len(prds))
	for i, prd := range prds {
		app[i] = toAppProduct(prd)
	}

	return app
}
