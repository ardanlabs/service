package vproductdb

import (
	"time"

	"github.com/ardanlabs/service/business/domain/vproductbus"
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

func toBusProduct(dbPrd dbProduct) vproductbus.Product {
	bus := vproductbus.Product{
		ID:          dbPrd.ID,
		UserID:      dbPrd.UserID,
		Name:        dbPrd.Name,
		Cost:        dbPrd.Cost,
		Quantity:    dbPrd.Quantity,
		DateCreated: dbPrd.DateCreated.In(time.Local),
		DateUpdated: dbPrd.DateUpdated.In(time.Local),
		UserName:    dbPrd.UserName,
	}

	return bus
}

func toBusProducts(dbPrds []dbProduct) []vproductbus.Product {
	bus := make([]vproductbus.Product, len(dbPrds))

	for i, dbPrd := range dbPrds {
		bus[i] = toBusProduct(dbPrd)
	}

	return bus
}
