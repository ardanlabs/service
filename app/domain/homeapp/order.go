package homeapp

import (
	"github.com/ardanlabs/service/business/api/order"
	"github.com/ardanlabs/service/business/domain/homebus"
)

var defaultOrderBy = order.NewBy("home_id", order.ASC)

var orderByFields = map[string]string{
	"home_id": homebus.OrderByID,
	"type":    homebus.OrderByType,
	"user_id": homebus.OrderByUserID,
}
