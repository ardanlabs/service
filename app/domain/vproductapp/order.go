package vproductapp

import (
	"github.com/ardanlabs/service/business/api/order"
	"github.com/ardanlabs/service/business/domain/vproductbus"
)

const (
	orderByProductID = "product_id"
	orderByUserID    = "user_id"
	orderByName      = "name"
	orderByCost      = "cost"
	orderByQuantity  = "quantity"
	orderByUserName  = "user_name"
)

var orderByFields = map[string]string{
	orderByProductID: vproductbus.OrderByProductID,
	orderByUserID:    vproductbus.OrderByUserID,
	orderByName:      vproductbus.OrderByName,
	orderByCost:      vproductbus.OrderByCost,
	orderByQuantity:  vproductbus.OrderByQuantity,
	orderByUserName:  vproductbus.OrderByUserName,
}

func parseOrder(qp QueryParams) (order.By, error) {
	return order.Parse(orderByFields, qp.OrderBy, order.NewBy(orderByProductID, order.ASC))
}
