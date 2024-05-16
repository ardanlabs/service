package vproductbus

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
	UserName *string
}

// Validate can perform a check of the data against the validate tags.
func (qf *QueryFilter) Validate() error {
	if err := validate.Check(qf); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}
