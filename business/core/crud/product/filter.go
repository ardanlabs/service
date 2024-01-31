package product

import (
	"fmt"

	"github.com/ardanlabs/service/foundation/validate"
	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ID       *uuid.UUID
	Name     *string `validate:"omitempty,min=3"`
	Cost     *float64
	Quantity *int
}

// Validate can perform a check of the data against the validate tags.
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
