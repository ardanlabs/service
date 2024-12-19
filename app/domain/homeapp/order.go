package homeapp

import (
	"github.com/ardanlabs/service/business/domain/homebus"
)

var orderByFields = map[string]string{
	"home_id": homebus.OrderByID,
	"type":    homebus.OrderByType,
	"user_id": homebus.OrderByUserID,
}
