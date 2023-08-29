package homedb

import (
	"fmt"

	"github.com/ardanlabs/service/business/data/order"
)

var orderByFields = map[string]string{}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
