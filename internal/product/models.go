package product

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// Product is an item we sell.
type Product struct {
	ID           bson.ObjectId `bson:"_id" json:"id"`                      // Unique identifier.
	Name         string        `bson:"name" json:"name"`                   // Display name of the product.
	Cost         int           `bson:"cost" json:"cost"`                   // Price for one item in cents.
	Quantity     int           `bson:"quantity" json:"quantity"`           // Original number of items available.
	DateCreated  time.Time     `bson:"date_created" json:"date_created"`   // When the product was added.
	DateModified time.Time     `bson:"date_modified" json:"date_modified"` // When the product record was lost modified.
}

// NewProduct is what we require from clients when adding a Product.
type NewProduct struct {
	Name     string `json:"name" validate:"required"`
	Cost     int    `json:"cost" validate:"required,gte=0"`
	Quantity int    `json:"quantity" validate:"required,gte=1"`
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

// Sale represents a transaction where we sold some quantity of a
// Product.
type Sale struct{}

// NewSale defines what we require when creating a Sale record.
type NewSale struct{}
