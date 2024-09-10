package vproductbus

import (
	"time"

	"github.com/ardanlabs/service/business/types/name"
	"github.com/google/uuid"
)

// Product represents an individual product with extended information.
type Product struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        name.Name
	Cost        float64
	Quantity    int
	DateCreated time.Time
	DateUpdated time.Time
	UserName    name.Name
}
