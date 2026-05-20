package commondb

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ardanlabs/service/business/domain/productbus"
)

// ApplyFilter appends a WHERE clause to buf based on the fields set in
// filter and adds the matching named bind values to data. The clause uses
// only portable SQL so it works for every supported engine.
func ApplyFilter(filter productbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["product_id"] = filter.ID
		wc = append(wc, "product_id = :product_id")
	}

	if filter.Name != nil {
		data["name"] = fmt.Sprintf("%%%s%%", filter.Name)
		wc = append(wc, "name LIKE :name")
	}

	if filter.Cost != nil {
		data["cost"] = filter.Cost
		wc = append(wc, "cost = :cost")
	}

	if filter.Quantity != nil {
		data["quantity"] = filter.Quantity
		wc = append(wc, "quantity = :quantity")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
