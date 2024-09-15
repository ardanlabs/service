package productbus

import (
	"time"

	"github.com/ardanlabs/service/business/types/money"
	"github.com/ardanlabs/service/business/types/name"
	"github.com/ardanlabs/service/business/types/quantity"
	"github.com/google/uuid"
)

// Product represents an individual product.
type Product struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        name.Name
	Cost        money.Money
	Quantity    quantity.Quantity
	DateCreated time.Time
	DateUpdated time.Time
}

// NewProduct is what we require from clients when adding a Product.
type NewProduct struct {
	UserID   uuid.UUID
	Name     name.Name
	Cost     money.Money
	Quantity quantity.Quantity
}

// UpdateProduct defines what information may be provided to modify an
// existing Product. All fields are optional so clients can send just the
// fields they want changed. It uses pointer fields so we can differentiate
// between a field that was not provided and a field that was provided as
// explicitly blank. Normally we do not want to use pointers to basic types but
// we make exceptions around marshalling/unmarshalling.
type UpdateProduct struct {
	Name     *name.Name
	Cost     *money.Money
	Quantity *quantity.Quantity
}
