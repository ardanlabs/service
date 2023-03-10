package usergrp

import (
	"errors"
	"net/http"

	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/order"
	"github.com/ardanlabs/service/business/sys/validate"
)

var orderByFields = map[string]struct{}{
	user.OrderByID:      {},
	user.OrderByName:    {},
	user.OrderByEmail:   {},
	user.OrderByRoles:   {},
	user.OrderByEnabled: {},
}

func parseOrder(r *http.Request) (order.By, error) {
	orderBy, err := order.Parse(r, user.DefaultOrderBy)
	if err != nil {
		return order.By{}, err
	}

	if _, exists := orderByFields[orderBy.Field]; !exists {
		return order.By{}, validate.NewFieldsError(orderBy.Field, errors.New("order field does not exist"))
	}

	return orderBy, nil
}
