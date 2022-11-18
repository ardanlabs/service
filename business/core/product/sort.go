package product

import "github.com/ardanlabs/service/business/data/sort"

// Order provides acces to ordering functionality.
var Order = sort.NewOrder(orderByfields, OrderByID)

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = sort.OrderBy{Field: OrderByID, Direction: sort.ASC}

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

// NewOrderBy creates a new OrderBy with field validation.
func NewOrderBy(field string, direction string) (sort.OrderBy, error) {
	return Order.NewOrderBy(field, direction)
}
