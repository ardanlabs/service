package product_test

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/ardanlabs/service/app/domain/productapp"
	"github.com/ardanlabs/service/app/sdk/apitest"
	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/app/sdk/query"
	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/google/go-cmp/cmp"
)

func query200(sd apitest.SeedData) []apitest.Table {
	prds := make([]productbus.Product, 0, len(sd.Admins[0].Products)+len(sd.Users[0].Products))
	prds = append(prds, sd.Admins[0].Products...)
	prds = append(prds, sd.Users[0].Products...)

	sort.Slice(prds, func(i, j int) bool {
		return prds[i].ID.String() <= prds[j].ID.String()
	})

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/products?page=1&rows=10&orderBy=product_id,ASC",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[productapp.Product]{},
			ExpResp: &query.Result[productapp.Product]{
				Page:        1,
				RowsPerPage: 10,
				Total:       len(prds),
				Items:       toAppProducts(prds),
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func query400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "bad-query-filter",
			URL:        "/v1/products?page=1&rows=10&name=$#!",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "[{\"field\":\"name\",\"error\":\"invalid name \\\"$#!\\\"\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-orderby-value",
			URL:        "/v1/products?page=1&rows=10&orderBy=roduct_id,ASC",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "[{\"field\":\"order\",\"error\":\"unknown order: roduct_id\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/products/%s", sd.Users[0].Products[0].ID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &productapp.Product{},
			ExpResp:    toAppProductPtr(sd.Users[0].Products[0]),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
