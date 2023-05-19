package summarydb

import (
	"fmt"

	"github.com/ardanlabs/service/business/cview/user/summary"
	"github.com/ardanlabs/service/business/data/order"
)

var orderByFields = map[string]string{
	summary.OrderByUserID:   "user_id",
	summary.OrderByUserName: "user_name",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
