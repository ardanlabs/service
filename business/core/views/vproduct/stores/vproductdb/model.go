package vproductdb

import (
	"time"

	"github.com/ardanlabs/service/business/core/views/vproduct"
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
	UserName    string    `db:"user_name"`
}

func toCoreProduct(dbPrd dbProduct) vproduct.Product {
	prd := vproduct.Product{
		ID:          dbPrd.ID,
		UserID:      dbPrd.UserID,
		Name:        dbPrd.Name,
		Cost:        dbPrd.Cost,
		Quantity:    dbPrd.Quantity,
		DateCreated: dbPrd.DateCreated.In(time.Local),
		DateUpdated: dbPrd.DateUpdated.In(time.Local),
		UserName:    dbPrd.UserName,
	}

	return prd
}

func toCoreProducts(dbPrds []dbProduct) []vproduct.Product {
	prds := make([]vproduct.Product, len(dbPrds))

	for i, dbPrd := range dbPrds {
		prds[i] = toCoreProduct(dbPrd)
	}

	return prds
}
