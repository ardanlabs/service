package vproductdb

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ardanlabs/service/business/domain/vproductbus"
)

func (s *Store) applyFilter(filter vproductbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["product_id"] = *filter.ID
		wc = append(wc, "product_id = :product_id")
	}

	if filter.Name != nil {
		data["name"] = fmt.Sprintf("%%%s%%", *filter.Name)
		wc = append(wc, "name LIKE :name")
	}

	if filter.Cost != nil {
		data["cost"] = *filter.Cost
		wc = append(wc, "cost = :cost")
	}

	if filter.Quantity != nil {
		data["quantity"] = *filter.Quantity
		wc = append(wc, "quantity = :quantity")
	}

	if filter.UserName != nil {
		data["user_name"] = fmt.Sprintf("%%%s%%", *filter.Name)
		wc = append(wc, "user_name LIKE :user_name")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
