package productdb

import (
	"time"

	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/google/uuid"
)

type dbProduct struct {
	ID          uuid.UUID `db:"product_id"`
	UserID      uuid.UUID `db:"user_id"`
	Name        string    `db:"name"`
	Cost        float64   `db:"cost"`
	Quantity    int       `db:"quantity"`
	DateCreated time.Time `db:"date_created"`
	DateUpdated time.Time `db:"date_updated"`
}

func toDBProduct(prd product.Product) dbProduct {
	prdDB := dbProduct{
		ID:          prd.ID,
		UserID:      prd.UserID,
		Name:        prd.Name,
		Cost:        prd.Cost,
		Quantity:    prd.Quantity,
		DateCreated: prd.DateCreated.UTC(),
		DateUpdated: prd.DateUpdated.UTC(),
	}

	return prdDB
}

func toCoreProduct(dbPrd dbProduct) product.Product {
	prd := product.Product{
		ID:          dbPrd.ID,
		UserID:      dbPrd.UserID,
		Name:        dbPrd.Name,
		Cost:        dbPrd.Cost,
		Quantity:    dbPrd.Quantity,
		DateCreated: dbPrd.DateCreated.In(time.Local),
		DateUpdated: dbPrd.DateUpdated.In(time.Local),
	}

	return prd
}

func toCoreProducts(dbPrds []dbProduct) []product.Product {
	prds := make([]product.Product, len(dbPrds))

	for i, dbPrd := range dbPrds {
		prds[i] = toCoreProduct(dbPrd)
	}

	return prds
}
