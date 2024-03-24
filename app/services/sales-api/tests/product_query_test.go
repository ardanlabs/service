package tests

import (
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/crud/productgrp"
	"github.com/ardanlabs/service/business/web/page"
	"github.com/google/go-cmp/cmp"
)

func productQuery200(sd seedData) []tableData {
	total := len(sd.admins[0].products) + len(sd.users[0].products)

	table := []tableData{
		{
			name:       "basic",
			url:        "/v1/products?page=1&rows=10&orderBy=product_id,DESC",
			token:      sd.admins[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &page.Document[productgrp.AppProduct]{},
			expResp: &page.Document[productgrp.AppProduct]{
				Page:        1,
				RowsPerPage: 10,
				Total:       total,
				Items:       toAppProducts(append(sd.admins[0].products, sd.users[0].products...)),
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				resp := x.(*page.Document[productgrp.AppProduct])
				exp := y.(*page.Document[productgrp.AppProduct])

				var found int
				for _, r := range resp.Items {
					for _, e := range exp.Items {
						if e.ID == r.ID {
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

func productQueryByID200(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "basic",
			url:        fmt.Sprintf("/v1/products/%s", sd.users[0].products[0].ID),
			token:      sd.users[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &productgrp.AppProduct{},
			expResp:    toAppProductPtr(sd.users[0].products[0]),
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}
