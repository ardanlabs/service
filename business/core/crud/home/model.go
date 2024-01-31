package home

import (
	"time"

	"github.com/google/uuid"
)

// Address represents an individual address.
type Address struct {
	Address1 string
	Address2 string
	ZipCode  string
	City     string
	State    string
	Country  string
}

// Home represents an individual home.
type Home struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Type        Type
	Address     Address
	DateCreated time.Time
	DateUpdated time.Time
}

// NewHome is what we require from clients when adding a Home.
type NewHome struct {
	UserID  uuid.UUID
	Type    Type
	Address Address
}

// UpdateAddress is what fields can be updated in the store.
type UpdateAddress struct {
	Address1 *string
	Address2 *string
	ZipCode  *string
	City     *string
	State    *string
	Country  *string
}

// UpdateHome defines what informaton may be provided to modify an existing
// Home. All fields are optional so clients can send only the fields they want
// changed. It uses pointer fields so we can differentiate between a field that
// was not provided and a field that was provided as explicity blank. Normally
// we do not want to use pointers to basic types but we make exepction around
// marshalling/unmarshalling.
type UpdateHome struct {
	Type    *Type
	Address *UpdateAddress
}
