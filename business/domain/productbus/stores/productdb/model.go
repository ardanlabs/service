package productdb

import (
	"fmt"
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

func toDBProduct(bus productbus.Product) product {
	db := product{
		ID:          bus.ID,
		UserID:      bus.UserID,
		Name:        bus.Name.String(),
		Cost:        bus.Cost,
		Quantity:    bus.Quantity,
		DateCreated: bus.DateCreated.UTC(),
		DateUpdated: bus.DateUpdated.UTC(),
	}

	return db
}

func toBusProduct(db product) (productbus.Product, error) {
	name, err := productbus.Names.Parse(db.Name)
	if err != nil {
		return productbus.Product{}, fmt.Errorf("parse name: %w", err)
	}

	bus := productbus.Product{
		ID:          db.ID,
		UserID:      db.UserID,
		Name:        name,
		Cost:        db.Cost,
		Quantity:    db.Quantity,
		DateCreated: db.DateCreated.In(time.Local),
		DateUpdated: db.DateUpdated.In(time.Local),
	}

	return bus, nil
}

func toBusProducts(dbs []product) ([]productbus.Product, error) {
	bus := make([]productbus.Product, len(dbs))

	for i, db := range dbs {
		var err error
		bus[i], err = toBusProduct(db)
		if err != nil {
			return nil, err
		}
	}

	return bus, nil
}
