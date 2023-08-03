package usersummarygrp

import (
	"errors"
	"net/http"

	"github.com/ardanlabs/service/business/core/usersummary"
	"github.com/ardanlabs/service/business/data/order"
	"github.com/ardanlabs/service/foundation/validate"
)

var orderByFields = map[string]struct{}{
	usersummary.OrderByUserID:   {},
	usersummary.OrderByUserName: {},
}

func parseOrder(r *http.Request) (order.By, error) {
	orderBy, err := order.Parse(r, usersummary.DefaultOrderBy)
	if err != nil {
		return order.By{}, err
	}

	if _, exists := orderByFields[orderBy.Field]; !exists {
		return order.By{}, validate.NewFieldsError(orderBy.Field, errors.New("order field does not exist"))
	}

	return orderBy, nil
}
