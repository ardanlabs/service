package tests

import (
	"net/http"
	"sort"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/views/vproductgrp"
	"github.com/ardanlabs/service/business/web/page"
	"github.com/google/go-cmp/cmp"
)

func vproductQuery200(sd seedData) []tableData {
	prds := toAppVProducts(sd.admins[0].User, sd.admins[0].products)
	prds = append(prds, toAppVProducts(sd.users[0].User, sd.users[0].products)...)

	sort.Slice(prds, func(i, j int) bool {
		return prds[i].ID <= prds[j].ID
	})

	table := []tableData{
		{
			name:       "basic",
			url:        "/v1/vproducts?page=1&rows=10&orderBy=product_id,ASC",
			token:      sd.admins[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &page.Document[vproductgrp.AppProduct]{},
			expResp: &page.Document[vproductgrp.AppProduct]{
				Page:        1,
				RowsPerPage: 10,
				Total:       len(prds),
				Items:       prds,
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
