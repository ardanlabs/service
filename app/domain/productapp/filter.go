package productapp

import (
	"strconv"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/google/uuid"
)

func parseFilter(qp QueryParams) (productbus.QueryFilter, error) {
	var filter productbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return productbus.QueryFilter{}, errs.NewFieldsError("product_id", err)
		}
		filter.ID = &id
	}

	if qp.Name != "" {
		name, err := productbus.ParseName(qp.Name)
		if err != nil {
			return productbus.QueryFilter{}, errs.NewFieldsError("name", err)
		}
		filter.Name = &name
	}

	if qp.Cost != "" {
		cst, err := strconv.ParseFloat(qp.Cost, 64)
		if err != nil {
			return productbus.QueryFilter{}, errs.NewFieldsError("cost", err)
		}
		filter.Cost = &cst
	}

	if qp.Quantity != "" {
		qua, err := strconv.ParseInt(qp.Quantity, 10, 64)
		if err != nil {
			return productbus.QueryFilter{}, errs.NewFieldsError("quantity", err)
		}
		i := int(qua)
		filter.Quantity = &i
	}

	return filter, nil
}
