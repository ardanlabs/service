package productgrp

import (
	"errors"
	"net/http"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/data/order"
	"github.com/ardanlabs/service/business/sys/validate"
)

var orderByFields = map[string]struct{}{
	product.OrderByProdID:   {},
	product.OrderByName:     {},
	product.OrderByCost:     {},
	product.OrderByQuantity: {},
	product.OrderBySold:     {},
	product.OrderByRevenue:  {},
	product.OrderByUserID:   {},
}

func parseOrder(r *http.Request) (order.By, error) {
	orderBy, err := order.Parse(r, product.DefaultOrderBy)
	if err != nil {
		return order.By{}, err
	}

	if _, exists := orderByFields[orderBy.Field]; !exists {
		return order.By{}, validate.NewFieldsError(orderBy.Field, errors.New("order field does not exist"))
	}

	return orderBy, nil
}
