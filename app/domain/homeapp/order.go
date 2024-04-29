package homeapp

import (
	"github.com/ardanlabs/service/business/api/order"
	"github.com/ardanlabs/service/business/domain/homebus"
)

const (
	orderByID     = "home_id"
	orderByType   = "type"
	orderByUserID = "user_id"
)

var orderByFields = map[string]string{
	orderByID:     homebus.OrderByID,
	orderByType:   homebus.OrderByType,
	orderByUserID: homebus.OrderByUserID,
}

func parseOrder(qp QueryParams) (order.By, error) {
	return order.Parse(orderByFields, qp.OrderBy, order.NewBy(orderByID, order.ASC))
}
