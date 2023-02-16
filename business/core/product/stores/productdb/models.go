package productdb

import (
	"time"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/google/uuid"
)

// dbProduct represents an individual product.
type dbProduct struct {
	ID          uuid.UUID `db:"product_id"`   // Unique identifier.
	Name        string    `db:"name"`         // Display name of the product.
	Cost        int       `db:"cost"`         // Price for one item in cents.
	Quantity    int       `db:"quantity"`     // Original number of items available.
	Sold        int       `db:"sold"`         // Aggregate field showing number of items sold.
	Revenue     int       `db:"revenue"`      // Aggregate field showing total cost of sold items.
	UserID      uuid.UUID `db:"user_id"`      // ID of the user who created the product.
	DateCreated time.Time `db:"date_created"` // When the product was added.
	DateUpdated time.Time `db:"date_updated"` // When the product record was last modified.
}

// =============================================================================

func toDBProduct(prd product.Product) dbProduct {
	prdDB := dbProduct{
		ID:          prd.ID,
		Name:        prd.Name,
		Cost:        prd.Cost,
		Quantity:    prd.Quantity,
		Sold:        prd.Sold,
		Revenue:     prd.Revenue,
		UserID:      prd.UserID,
		DateCreated: prd.DateCreated.UTC(),
		DateUpdated: prd.DateUpdated.UTC(),
	}

	return prdDB
}

func toCoreProduct(dbPrd dbProduct) product.Product {
	prd := product.Product{
		ID:          dbPrd.ID,
		Name:        dbPrd.Name,
		Cost:        dbPrd.Cost,
		Quantity:    dbPrd.Quantity,
		Sold:        dbPrd.Sold,
		Revenue:     dbPrd.Revenue,
		UserID:      dbPrd.UserID,
		DateCreated: dbPrd.DateCreated.In(time.Local),
		DateUpdated: dbPrd.DateUpdated.In(time.Local),
	}

	return prd
}

func toCoreProductSlice(dbProducts []dbProduct) []product.Product {
	prds := make([]product.Product, len(dbProducts))
	for i, dbPrd := range dbProducts {
		prds[i] = toCoreProduct(dbPrd)
	}
	return prds
}
