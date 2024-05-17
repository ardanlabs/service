package vproduct_test

import (
	"net/http"
	"sort"

	"github.com/ardanlabs/service/api/sdk/http/apitest"
	"github.com/ardanlabs/service/app/domain/vproductapp"
	"github.com/ardanlabs/service/app/sdk/query"
	"github.com/google/go-cmp/cmp"
)

func query200(sd apitest.SeedData) []apitest.Table {
	prds := toAppVProducts(sd.Admins[0].User, sd.Admins[0].Products)
	prds = append(prds, toAppVProducts(sd.Users[0].User, sd.Users[0].Products)...)

	sort.Slice(prds, func(i, j int) bool {
		return prds[i].ID <= prds[j].ID
	})

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/vproducts?page=1&rows=10&orderBy=product_id,ASC",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[vproductapp.Product]{},
			ExpResp: &query.Result[vproductapp.Product]{
				Page:        1,
				RowsPerPage: 10,
				Total:       len(prds),
				Items:       prds,
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
