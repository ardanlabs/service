package vproductapp

import (
	"errors"

	"github.com/ardanlabs/service/business/api/order"
	"github.com/ardanlabs/service/business/domain/vproductbus"
	"github.com/ardanlabs/service/foundation/validate"
)

func parseOrder(qp QueryParams) (order.By, error) {
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

	orderBy, err := order.Parse(qp.OrderBy, order.NewBy(orderByProductID, order.ASC))
	if err != nil {
		return order.By{}, err
	}

	if _, exists := orderByFields[orderBy.Field]; !exists {
		return order.By{}, validate.NewFieldsError(orderBy.Field, errors.New("order field does not exist"))
	}

	orderBy.Field = orderByFields[orderBy.Field]

	return orderBy, nil
}
