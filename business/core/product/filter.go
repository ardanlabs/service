package product

import (
	"fmt"

	"github.com/ardanlabs/service/business/sys/validate"
	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
type QueryFilter struct {
	ID       *uuid.UUID `validate:"omitempty"`
	Name     *string    `validate:"omitempty,min=3"`
	Cost     *float64   `validate:"omitempty,numeric"`
	Quantity *int       `validate:"omitempty,numeric"`
}

// Validate checks the data in the model is considered clean.
func (qf *QueryFilter) Validate() error {
	if err := validate.Check(qf); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// WithProductID sets the ID field of the QueryFilter value.
func (qf *QueryFilter) WithProductID(productID uuid.UUID) {
	qf.ID = &productID
}

// WithName sets the Name field of the QueryFilter value.
func (qf *QueryFilter) WithName(name string) {
	qf.Name = &name
}

// WithCost sets the Cost field of the QueryFilter value.
func (qf *QueryFilter) WithCost(cost float64) {
	qf.Cost = &cost
}

// WithQuantity sets the Quantity field of the QueryFilter value.
func (qf *QueryFilter) WithQuantity(quantity int) {
	qf.Quantity = &quantity
}
