package user

// QueryFilter holds the available fields filters to search
// for schedules on the store.
type QueryFilter struct {
	ID    *string `validate:"omitempty,uuid4"`
	Name  *string `validate:"omitempty,min=3"`
	Email *string `validate:"omitempty,email"`
}

// ByID sets the ID field of the QueryFilter value.
func (f *QueryFilter) ByID(id string) {
	if id != "" {
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
func (f *QueryFilter) ByEmail(email string) {
	if email != "" {
		f.Email = &email
	}
}
