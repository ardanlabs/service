package user

import (
	"net/mail"
	"time"

	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics to support the api.
type QueryFilter struct {
	ID               *uuid.UUID    `validate:"omitempty"`
	Name             *string       `validate:"omitempty,min=3"`
	Email            *mail.Address `validate:"omitempty"`
	StartCreatedDate *time.Time    `validate:"omitempty"`
	EndCreatedDate   *time.Time    `validate:"omitempty"`
}

// WithUserID sets the ID field of the QueryFilter value.
func (qf *QueryFilter) WithUserID(userID uuid.UUID) {
	qf.ID = &userID
}

// WithName sets the Name field of the QueryFilter value.
func (qf *QueryFilter) WithName(name string) {
	qf.Name = &name
}

// WithEmail sets the Email field of the QueryFilter value.
func (qf *QueryFilter) WithEmail(email mail.Address) {
	qf.Email = &email
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
