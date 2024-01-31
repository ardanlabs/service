package product

import "github.com/ardanlabs/service/business/web/v1/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByProductID, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByProductID = "product_id"
	OrderByUserID    = "user_id"
	OrderByName      = "name"
	OrderByCost      = "cost"
	OrderByQuantity  = "quantity"
)
