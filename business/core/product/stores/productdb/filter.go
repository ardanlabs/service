package productdb

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ardanlabs/service/business/core/product"
)

func (s *Store) applyFilter(filter product.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["product_id"] = *filter.ID
		wc = append(wc, "p.product_id = :product_id")
	}

	if filter.Name != nil {
		data["name"] = fmt.Sprintf("%%%s%%", *filter.Name)
		wc = append(wc, "p.name LIKE :name")
	}

	if filter.Cost != nil {
		data["cost"] = *filter.Cost
		wc = append(wc, "p.cost = :cost")
	}

	if filter.Quantity != nil {
		data["quantity"] = *filter.Quantity
		wc = append(wc, "p.quantity = :quantity")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
