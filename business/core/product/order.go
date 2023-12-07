package product

import "github.com/ardanlabs/service/business/web/v1/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByProdID, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByProdID   = "product_id"
	OrderByName     = "name"
	OrderByCost     = "cost"
	OrderByQuantity = "quantity"
	OrderBySold     = "sold"
	OrderByRevenue  = "revenue"
	OrderByUserID   = "user_id"
)
