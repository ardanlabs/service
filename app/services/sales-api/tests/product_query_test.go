package tests

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/crud/productgrp"
	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/web/page"
	"github.com/google/go-cmp/cmp"
)

func productQuery200(sd seedData) []tableData {
	prds := make([]product.Product, 0, len(sd.admins[0].products)+len(sd.users[0].products))
	prds = append(prds, sd.admins[0].products...)
	prds = append(prds, sd.users[0].products...)

	sort.Slice(prds, func(i, j int) bool {
		return prds[i].ID.String() <= prds[j].ID.String()
	})

	table := []tableData{
		{
			name:       "basic",
			url:        "/v1/products?page=1&rows=10&orderBy=product_id,ASC",
			token:      sd.admins[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &page.Document[productgrp.AppProduct]{},
			expResp: &page.Document[productgrp.AppProduct]{
				Page:        1,
				RowsPerPage: 10,
				Total:       len(prds),
				Items:       toAppProducts(prds),
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
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
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
