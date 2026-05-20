package commondb

import (
	"fmt"

	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/ardanlabs/service/business/sdk/order"
)

// OrderByFields maps a productbus ordering key to its database column name.
// The mapping is shared because every engine uses the same column names.
var OrderByFields = map[string]string{
	productbus.OrderByProductID: "product_id",
	productbus.OrderByUserID:    "user_id",
	productbus.OrderByName:      "name",
	productbus.OrderByCost:      "cost",
	productbus.OrderByQuantity:  "quantity",
}

// OrderByClause builds an ORDER BY clause for the given ordering. It
// returns an error if the field is not in OrderByFields.
func OrderByClause(orderBy order.By) (string, error) {
	by, exists := OrderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
