package product

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// Product is something we have for sale.
type Product struct {
	ID           bson.ObjectId `bson:"_id" json:"id"`                      // Unique identifier.
	Name         string        `bson:"name" json:"name"`                   // Display name of the product.
	Notes        string        `bson:"notes" json:"notes"`                 // Optional descriptive field.
	Family       string        `bson:"family" json:"family"`               // Which family provided the product.
	UnitPrice    int           `bson:"unit_price" json:"unit_price"`       // Price for one item in cents.
	Quantity     int           `bson:"quantity" json:"quantity"`           // Original number of items available.
	DateCreated  time.Time     `bson:"date_created" json:"date_created"`   // When the product was added.
	DateModified time.Time     `bson:"date_modified" json:"date_modified"` // When the product record was lost modified.
}

// NewProduct defines the information we need when adding a Product to
// our offerings.
type NewProduct struct {
	Name      string `json:"name" validate:"required"`
	Notes     string `json:"notes"`
	Family    string `json:"family" validate:"required"`
	UnitPrice int    `json:"unit_price" validate:"required,gte=0"`
	Quantity  int    `json:"quantity" validate:"required,gte=1"`
}

// UpdateProduct defines what information may be provided to modify an
// existing Product. All fields are optional so clients can send just the
// fields they want changed. It uses pointer fields so we can differentiate
// between a field that was not provided and a field that was provided as
// explicitly blank. Normally we do not want to use pointers to basic types but
// we make exceptions around marshalling/unmarshalling.
type UpdateProduct struct {
	Name      *string `json:"name"`
	Notes     *string `json:"notes"`
	Family    *string `json:"family"`
	UnitPrice *int    `json:"unit_price" validate:"omitempty,gte=0"`
	Quantity  *int    `json:"quantity" validate:"omitempty,gte=1"`
}

// Sale represents a transaction where we sold some quantity of a
// Product.
type Sale struct{}

// NewSale defines what we require when creating a Sale record.
type NewSale struct{}
