package product

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/ardanlabs/service/business/sys/validate"
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

// Validate checks the data in the model is considered clean.
func (qf *QueryFilter) Validate() error {
	if err := validate.Check(qf); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// WithProductID sets the ID field of the QueryFilter value.
func (qf *QueryFilter) WithProductID(productID uuid.UUID) {
	var zero uuid.UUID
	if productID != zero {
		qf.ID = &productID
	}
}

// WithName sets the Name field of the QueryFilter value.
func (qf *QueryFilter) WithName(name string) error {
	if name != "" {
		if !sqlInjection.MatchString(name) {
			return errors.New("invalid name format")
		}

		qf.Name = &name
	}

	return nil
}

// WithCost sets the Cost field of the QueryFilter value.
func (qf *QueryFilter) WithCost(cost int) {
	qf.Cost = &cost
}

// WithQuantity sets the Quantity field of the QueryFilter value.
func (qf *QueryFilter) WithQuantity(quantity int) {
	qf.Quantity = &quantity
}
