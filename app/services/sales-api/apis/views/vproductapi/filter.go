package vproductapi

import (
	"strconv"

	"github.com/ardanlabs/service/business/core/views/vproduct"
	"github.com/ardanlabs/service/foundation/validate"
	"github.com/google/uuid"
)

func parseFilter(qp QueryParams) (vproduct.QueryFilter, error) {
	var filter vproduct.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return vproduct.QueryFilter{}, validate.NewFieldsError("product_id", err)
		}
		filter.WithID(id)
	}

	if qp.Name != "" {
		filter.WithName(qp.Name)
	}

	if qp.Cost != "" {
		cst, err := strconv.ParseFloat(qp.Cost, 64)
		if err != nil {
			return vproduct.QueryFilter{}, validate.NewFieldsError("cost", err)
		}
		filter.WithCost(cst)
	}

	if qp.Quantity != "" {
		qua, err := strconv.ParseInt(qp.Quantity, 10, 64)
		if err != nil {
			return vproduct.QueryFilter{}, validate.NewFieldsError("quantity", err)
		}
		filter.WithQuantity(int(qua))
	}

	if qp.UserName != "" {
		filter.WithUserName(qp.UserName)
	}

	return filter, nil
}
