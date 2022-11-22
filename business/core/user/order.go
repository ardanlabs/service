package user

import "github.com/ardanlabs/service/business/data/order"

var ordering = order.New(orderByfields, OrderByID)

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.By{Field: OrderByID, Direction: order.ASC}

// Set of fields that the results can be ordered by. These are the names
// that should be used by the application layer.
const (
	OrderByID      = "id"
	OrderByName    = "name"
	OrderByEmail   = "email"
	OrderByRoles   = "roles"
	OrderByEnabled = "enabled"
)

// orderByfields is the map of fields that is used to perform validation.
var orderByfields = map[string]bool{
	OrderByID:      true,
	OrderByName:    true,
	OrderByEmail:   true,
	OrderByRoles:   true,
	OrderByEnabled: true,
}

// NewOrderBy creates a new order.By with field validation.
func NewOrderBy(field string, direction string) (order.By, error) {
	return ordering.By(field, direction)
}

// ParseOrderBy constructs an order.By value by parsing a string in the form
// of "field,direction".
func ParseOrderBy(query string) (order.By, error) {
	return ordering.ParseOrderBy(query)
}
