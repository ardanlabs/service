package productapp

import (
	"strconv"

	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/foundation/validate"
	"github.com/google/uuid"
)

func parseFilter(qp QueryParams) (product.QueryFilter, error) {
	var filter product.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return product.QueryFilter{}, validate.NewFieldsError("product_id", err)
		}
		filter.WithProductID(id)
	}

	if qp.Name != "" {
		filter.WithName(qp.Name)
	}

	if qp.Cost != "" {
		cst, err := strconv.ParseFloat(qp.Cost, 64)
		if err != nil {
			return product.QueryFilter{}, validate.NewFieldsError("cost", err)
		}
		filter.WithCost(cst)
	}

	if qp.Quantity != "" {
		qua, err := strconv.ParseInt(qp.Quantity, 10, 64)
		if err != nil {
			return product.QueryFilter{}, validate.NewFieldsError("quantity", err)
		}
		filter.WithQuantity(int(qua))
	}

	return filter, nil
}
