package productdb

import (
	"time"

	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/google/uuid"
)

type product struct {
	ID          uuid.UUID `db:"product_id"`
	UserID      uuid.UUID `db:"user_id"`
	Name        string    `db:"name"`
	Cost        float64   `db:"cost"`
	Quantity    int       `db:"quantity"`
	DateCreated time.Time `db:"date_created"`
	DateUpdated time.Time `db:"date_updated"`
}

func toDBProduct(prd productbus.Product) product {
	db := product{
		ID:          prd.ID,
		UserID:      prd.UserID,
		Name:        prd.Name,
		Cost:        prd.Cost,
		Quantity:    prd.Quantity,
		DateCreated: prd.DateCreated.UTC(),
		DateUpdated: prd.DateUpdated.UTC(),
	}

	return db
}

func toBusProduct(dbPrd product) productbus.Product {
	bus := productbus.Product{
		ID:          dbPrd.ID,
		UserID:      dbPrd.UserID,
		Name:        dbPrd.Name,
		Cost:        dbPrd.Cost,
		Quantity:    dbPrd.Quantity,
		DateCreated: dbPrd.DateCreated.In(time.Local),
		DateUpdated: dbPrd.DateUpdated.In(time.Local),
	}

	return bus
}

func toBusProducts(dbPrds []product) []productbus.Product {
	bus := make([]productbus.Product, len(dbPrds))

	for i, dbPrd := range dbPrds {
		bus[i] = toBusProduct(dbPrd)
	}

	return bus
}
