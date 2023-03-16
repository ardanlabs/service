package user

import "github.com/ardanlabs/service/business/data/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

// Set of fields that the results can be ordered by. These are the names
// that should be used by the application layer.
const (
	OrderByID      = "userid"
	OrderByName    = "name"
	OrderByEmail   = "email"
	OrderByRoles   = "roles"
	OrderByEnabled = "enabled"
)
