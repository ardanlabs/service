// Package commondb provides the engine-agnostic pieces every product SQL
// store needs: the database row struct, the conversions between the row
// struct and the business model, the WHERE clause builder, and the
// order-by field map. Engine-specific store packages (productpg,
// productsqlite, ...) own the SQL statements themselves and delegate to
// the helpers in this package for the parts that do not vary.
package commondb

import (
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/ardanlabs/service/business/types/money"
	"github.com/ardanlabs/service/business/types/name"
	"github.com/ardanlabs/service/business/types/quantity"
	"github.com/google/uuid"
)

// ProductDB is the database representation of a product row. It is shared
// by every SQL engine implementation because all of them use the same
// column set and types.
type ProductDB struct {
	ID          uuid.UUID `db:"product_id"`
	UserID      uuid.UUID `db:"user_id"`
	Name        string    `db:"name"`
	Cost        float64   `db:"cost"`
	Quantity    int       `db:"quantity"`
	DateCreated time.Time `db:"date_created"`
	DateUpdated time.Time `db:"date_updated"`
}

// ToDBProduct converts a business product into its database representation.
func ToDBProduct(bus productbus.Product) ProductDB {
	db := ProductDB{
		ID:          bus.ID,
		UserID:      bus.UserID,
		Name:        bus.Name.String(),
		Cost:        bus.Cost.Value(),
		Quantity:    bus.Quantity.Value(),
		DateCreated: bus.DateCreated.UTC(),
		DateUpdated: bus.DateUpdated.UTC(),
	}

	return db
}

// ToBusProduct converts a database row into a business product, parsing
// and validating the typed fields.
func ToBusProduct(db ProductDB) (productbus.Product, error) {
	name, err := name.Parse(db.Name)
	if err != nil {
		return productbus.Product{}, fmt.Errorf("parse name: %w", err)
	}

	cost, err := money.Parse(db.Cost)
	if err != nil {
		return productbus.Product{}, fmt.Errorf("parse cost: %w", err)
	}

	quantity, err := quantity.Parse(db.Quantity)
	if err != nil {
		return productbus.Product{}, fmt.Errorf("parse quantity: %w", err)
	}

	bus := productbus.Product{
		ID:          db.ID,
		UserID:      db.UserID,
		Name:        name,
		Cost:        cost,
		Quantity:    quantity,
		DateCreated: db.DateCreated.In(time.Local),
		DateUpdated: db.DateUpdated.In(time.Local),
	}

	return bus, nil
}

// ToBusProducts converts a slice of database rows into business products.
func ToBusProducts(dbs []ProductDB) ([]productbus.Product, error) {
	bus := make([]productbus.Product, len(dbs))

	for i, db := range dbs {
		var err error
		bus[i], err = ToBusProduct(db)
		if err != nil {
			return nil, err
		}
	}

	return bus, nil
}
