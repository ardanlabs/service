package product

import (
	"time"

	"github.com/google/uuid"
)

// Product represents an individual product.
type Product struct {
	ID          uuid.UUID
	Name        string
	Cost        float64
	Quantity    int
	Sold        int
	Revenue     int
	UserID      uuid.UUID
	DateCreated time.Time
	DateUpdated time.Time
}

// NewProduct is what we require from clients when adding a Product.
type NewProduct struct {
	Name     string
	Cost     float64
	Quantity int
	UserID   uuid.UUID
}

// UpdateProduct defines what information may be provided to modify an
// existing Product. All fields are optional so clients can send just the
// fields they want changed. It uses pointer fields so we can differentiate
// between a field that was not provided and a field that was provided as
// explicitly blank. Normally we do not want to use pointers to basic types but
// we make exceptions around marshalling/unmarshalling.
type UpdateProduct struct {
	Name     *string
	Cost     *float64
	Quantity *int
}
