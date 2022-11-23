package product

// QueryFilter holds the available fields filters to search
// for schedules on the store.
type QueryFilter struct {
	ID       *string `validate:"omitempty,uuid4"`
	Name     *string `validate:"omitempty,min=3"`
	Cost     *int    `validate:"omitempty,numeric"`
	Quantity *int    `validate:"omitempty,numeric"`
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

// ByCost sets the Cost field of the QueryFilter value.
func (f *QueryFilter) ByCost(cost int) {
	f.Cost = &cost
}

// ByQuantity sets the Quantity field of the QueryFilter value.
func (f *QueryFilter) ByQuantity(quantity int) {
	f.Quantity = &quantity
}
