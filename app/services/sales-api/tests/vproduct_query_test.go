package tests

import (
	"net/http"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/views/vproductgrp"
	"github.com/ardanlabs/service/business/web/page"
)

func vproductQuery200(sd seedData) []tableData {
	total := len(sd.admins[0].products) + len(sd.users[0].products)

	allProducts := toAppVProducts(sd.admins[0].User, sd.admins[0].products)
	allProducts = append(allProducts, toAppVProducts(sd.users[0].User, sd.users[0].products)...)

	table := []tableData{
		{
			name:       "basic",
			url:        "/v1/vproducts?page=1&rows=10&orderBy=product_id,DESC",
			token:      sd.admins[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &page.Document[vproductgrp.AppProduct]{},
			expResp: &page.Document[vproductgrp.AppProduct]{
				Page:        1,
				RowsPerPage: 10,
				Total:       total,
				Items:       allProducts,
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				resp := x.(*page.Document[vproductgrp.AppProduct])
				exp := y.(*page.Document[vproductgrp.AppProduct])

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
