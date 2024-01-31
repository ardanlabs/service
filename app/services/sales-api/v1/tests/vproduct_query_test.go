package tests

import (
	"net/http"

	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/vproductgrp"
	v1 "github.com/ardanlabs/service/business/web/v1"
)

func vproductQuery200(sd seedData) []tableData {
	total := len(sd.admins[1].products) + len(sd.users[1].products)

	allProducts := toAppVProducts(sd.admins[1].User, sd.admins[1].products)
	allProducts = append(allProducts, toAppVProducts(sd.users[1].User, sd.users[1].products)...)

	table := []tableData{
		{
			name:       "basic",
			url:        "/v1/vproducts?page=1&rows=10&orderBy=product_id,DESC",
			token:      sd.admins[1].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &v1.PageDocument[vproductgrp.AppProduct]{},
			expResp: &v1.PageDocument[vproductgrp.AppProduct]{
				Page:        1,
				RowsPerPage: 10,
				Total:       total,
				Items:       allProducts,
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				resp := x.(*v1.PageDocument[vproductgrp.AppProduct])
				exp := y.(*v1.PageDocument[vproductgrp.AppProduct])

				var found int
				for _, r := range resp.Items {
					for _, e := range exp.Items {
						if e.ID == r.ID && e.UserName == r.UserName {
							found++
							break
						}
					}
				}

				if found != total {
					return "number of expected products didn't match"
				}

				return ""
			},
		},
	}

	return table
}
