package home

import (
	"fmt"
	"time"

	"github.com/ardanlabs/service/foundation/validate"
	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ID               *uuid.UUID `validate:"omitempty"`
	UserID           *uuid.UUID `validate:"omitempty"`
	Type             *Type      `validate:"omitempty"`
	StartCreatedDate *time.Time `validate:"omitempty"`
	EndCreatedDate   *time.Time `validate:"omitempty"`
}

// Validate can perform a check of the data against the validate tags.
func (qf *QueryFilter) Validate() error {
	if err := validate.Check(qf); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

// WithHomeID sets the ID field of the QueryFilter value.
func (qf *QueryFilter) WithHomeID(homeID uuid.UUID) {
	qf.ID = &homeID
}

// WithUserID sets the ID field of the QueryFilter value.
func (qf *QueryFilter) WithUserID(userID uuid.UUID) {
	qf.UserID = &userID
}

// WithHomeType sets the Type field of the QueryFilter value.
func (qf *QueryFilter) WithHomeType(homeType Type) {
	qf.Type = &homeType
}

// WithStartDateCreated sets the DateCreated field of the QueryFilter value.
func (qf *QueryFilter) WithStartDateCreated(startDate time.Time) {
	d := startDate.UTC()
	qf.StartCreatedDate = &d
}

// WithEndCreatedDate sets the DateCreated field of the QueryFilter value.
func (qf *QueryFilter) WithEndCreatedDate(endDate time.Time) {
	d := endDate.UTC()
	qf.EndCreatedDate = &d
}
