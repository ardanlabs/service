package product

import (
	"errors"
	"regexp"

	"github.com/google/uuid"
)

// Used to check for sql injection problems.
var sqlInjection = regexp.MustCompile("^[A-Za-z0-9_]+$")

// =============================================================================

// QueryFilter holds the available fields a query can be filtered on.
type QueryFilter struct {
	ID       *uuid.UUID `validate:"omitempty,uuid4"`
	Name     *string    `validate:"omitempty,min=3"`
	Cost     *int       `validate:"omitempty,numeric"`
	Quantity *int       `validate:"omitempty,numeric"`
}

// ByID sets the ID field of the QueryFilter value.
func (f *QueryFilter) ByID(id uuid.UUID) {
	var zero uuid.UUID
	if id != zero {
		f.ID = &id
	}
}

// ByName sets the Name field of the QueryFilter value.
func (f *QueryFilter) ByName(name string) error {
	if name != "" {
		if !sqlInjection.MatchString(name) {
			return errors.New("invalid name format")
		}

		f.Name = &name
	}

	return nil
}

// ByCost sets the Cost field of the QueryFilter value.
func (f *QueryFilter) ByCost(cost int) {
	f.Cost = &cost
}

// ByQuantity sets the Quantity field of the QueryFilter value.
func (f *QueryFilter) ByQuantity(quantity int) {
	f.Quantity = &quantity
}
