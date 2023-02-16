package product

import (
	"time"

	"github.com/ardanlabs/service/business/sys/validate"
	"github.com/google/uuid"
)

// Product represents an individual product.
type Product struct {
	ID          uuid.UUID `json:"id"`          // Unique identifier.
	Name        string    `json:"name"`        // Display name of the product.
	Cost        int       `json:"cost"`        // Price for one item in cents.
	Quantity    int       `json:"quantity"`    // Original number of items available.
	Sold        int       `json:"sold"`        // Aggregate field showing number of items sold.
	Revenue     int       `json:"revenue"`     // Aggregate field showing total cost of sold items.
	UserID      uuid.UUID `json:"userID"`      // ID of the user who created the product.
	DateCreated time.Time `json:"dateCreated"` // When the product was added.
	DateUpdated time.Time `json:"dateUpdated"` // When the product record was last modified.
}

// NewProduct is what we require from clients when adding a Product.
type NewProduct struct {
	Name     string    `json:"name" validate:"required"`
	Cost     int       `json:"cost" validate:"required,gte=0"`
	Quantity int       `json:"quantity" validate:"gte=1"`
	UserID   uuid.UUID `json:"userID" validate:"required"`
}

// Validate checks the data in the model is considered clean.
func (np NewProduct) Validate() error {
	if err := validate.Check(np); err != nil {
		return err
	}
	return nil
}

// UpdateProduct defines what information may be provided to modify an
// existing Product. All fields are optional so clients can send just the
// fields they want changed. It uses pointer fields so we can differentiate
// between a field that was not provided and a field that was provided as
// explicitly blank. Normally we do not want to use pointers to basic types but
// we make exceptions around marshalling/unmarshalling.
type UpdateProduct struct {
	Name     *string `json:"name"`
	Cost     *int    `json:"cost" validate:"omitempty,gte=0"`
	Quantity *int    `json:"quantity" validate:"omitempty,gte=1"`
}

// Validate checks the data in the model is considered clean.
func (up UpdateProduct) Validate() error {
	if err := validate.Check(up); err != nil {
		return err
	}
	return nil
}
