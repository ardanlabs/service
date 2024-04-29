package productapp

import (
	"github.com/ardanlabs/service/business/api/order"
	"github.com/ardanlabs/service/business/domain/productbus"
)

const (
	orderByProductID = "product_id"
	orderByUserID    = "user_id"
	orderByName      = "name"
	orderByCost      = "cost"
	orderByQuantity  = "quantity"
)

var orderByFields = map[string]string{
	orderByProductID: productbus.OrderByProductID,
	orderByName:      productbus.OrderByName,
	orderByCost:      productbus.OrderByCost,
	orderByQuantity:  productbus.OrderByQuantity,
	orderByUserID:    productbus.OrderByUserID,
}

func parseOrder(qp QueryParams) (order.By, error) {
	return order.Parse(orderByFields, qp.OrderBy, order.NewBy(orderByProductID, order.ASC))
}
