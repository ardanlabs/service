package product

import (
	"github.com/ardanlabs/service/business/data/order"
)

var ordering = order.New(orderByfields, OrderByID)

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.By{Field: OrderByID, Direction: order.ASC}

// Set of fields that the results can be ordered by. These are the names
// that should be used by the application layer.
const (
	OrderByID       = "id"
	OrderByName     = "name"
	OrderByCost     = "cost"
	OrderByQuantity = "quanity"
	OrderBySold     = "sold"
	OrderByRevenue  = "revenue"
	OrderByUserID   = "userId"
)

// orderByfields is the map of fields that is used to perform validation.
var orderByfields = map[string]bool{
	OrderByID:       true,
	OrderByName:     true,
	OrderByCost:     true,
	OrderByQuantity: true,
	OrderBySold:     true,
	OrderByRevenue:  true,
	OrderByUserID:   true,
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
