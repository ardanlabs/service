package user

import (
	"fmt"
	"net/mail"
	"time"

	"github.com/ardanlabs/service/business/sys/validate"
	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
type QueryFilter struct {
	ID          *uuid.UUID    `validate:"omitempty,uuid4"`
	Name        *string       `validate:"omitempty,min=3"`
	Email       *mail.Address `validate:"omitempty,email"`
	DateCreated *time.Time
}

// Validate checks the data in the model is considered clean.
func (qf *QueryFilter) Validate() error {
	if err := validate.Check(qf); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// WithUserID sets the ID field of the QueryFilter value.
func (qf *QueryFilter) WithUserID(userID uuid.UUID) {
	qf.ID = &userID
}

// WithName sets the Name field of the QueryFilter value.
func (qf *QueryFilter) WithName(name string) {
	if name != "" {
		qf.Name = &name
	}
}

// WithEmail sets the Email field of the QueryFilter value.
func (qf *QueryFilter) WithEmail(email mail.Address) {
	qf.Email = &email
}

// WithDateCreated sets the DateCreated field of the QueryFilter value.
func (qf *QueryFilter) WithDateCreated(dateCreated time.Time) {
	d := dateCreated.UTC()
	qf.DateCreated = &d
}
