package vproduct

import (
	"time"

	"github.com/google/uuid"
)

// Product represents the extended data for a product.
type Product struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	Cost        float64
	Quantity    int
	DateCreated time.Time
	DateUpdated time.Time
	UserName    string
}
