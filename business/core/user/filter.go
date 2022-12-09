package user

import (
	"net/mail"

	"github.com/google/uuid"
)

// QueryFilter holds the available fields filters to search
// for schedules on the store.
type QueryFilter struct {
	ID    *uuid.UUID    `validate:"omitempty,uuid4"`
	Name  *string       `validate:"omitempty,min=3"`
	Email *mail.Address `validate:"omitempty,email"`
}

// ByID sets the ID field of the QueryFilter value.
func (f *QueryFilter) ByID(id uuid.UUID) {
	var zero uuid.UUID
	if id != zero {
		f.ID = &id
	}
}

// ByName sets the Name field of the QueryFilter value.
func (f *QueryFilter) ByName(name string) {
	if name != "" {
		f.Name = &name
	}
}

// ByEmail sets the Email field of the QueryFilter value.
func (f *QueryFilter) ByEmail(email mail.Address) {
	var zero mail.Address
	if email != zero {
		f.Email = &email
	}
}
