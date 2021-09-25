package product

import (
	"time"
	"unsafe"

	"github.com/ardanlabs/service/business/data/store/dbproduct"
)

// Product represents an individual product.
type Product struct {
	ID          string    `json:"id"`           // Unique identifier.
	Name        string    `json:"name"`         // Display name of the product.
	Cost        int       `json:"cost"`         // Price for one item in cents.
	Quantity    int       `json:"quantity"`     // Original number of items available.
	Sold        int       `json:"sold"`         // Aggregate field showing number of items sold.
	Revenue     int       `json:"revenue"`      // Aggregate field showing total cost of sold items.
	UserID      string    `json:"user_id"`      // ID of the user who created the product.
	DateCreated time.Time `json:"date_created"` // When the product was added.
	DateUpdated time.Time `json:"date_updated"` // When the product record was last modified.
}

// =============================================================================

func toProduct(dbPrd dbproduct.DBProduct) Product {
	pu := (*Product)(unsafe.Pointer(&dbPrd))
	return *pu
}

func toProductSlice(dbPrds []dbproduct.DBProduct) []Product {
	prds := make([]Product, len(dbPrds))
	for i, dbPrd := range dbPrds {
		prds[i] = toProduct(dbPrd)
	}
	return prds
}
