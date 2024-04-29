package userapp

import (
	"github.com/ardanlabs/service/business/api/order"
	"github.com/ardanlabs/service/business/domain/userbus"
)

const (
	orderByID      = "user_id"
	orderByName    = "name"
	orderByEmail   = "email"
	orderByRoles   = "roles"
	orderByEnabled = "enabled"
)

var orderByFields = map[string]string{
	orderByID:      userbus.OrderByID,
	orderByName:    userbus.OrderByName,
	orderByEmail:   userbus.OrderByEmail,
	orderByRoles:   userbus.OrderByRoles,
	orderByEnabled: userbus.OrderByEnabled,
}

func parseOrder(qp QueryParams) (order.By, error) {
	return order.Parse(orderByFields, qp.OrderBy, order.NewBy(orderByID, order.ASC))
}
