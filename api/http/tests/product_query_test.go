package tests

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/ardanlabs/service/app/api/page"
	"github.com/ardanlabs/service/app/domain/productapp"
	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/google/go-cmp/cmp"
)

func productQuery200(sd seedData) []table {
	prds := make([]productbus.Product, 0, len(sd.Admins[0].Products)+len(sd.Users[0].Products))
	prds = append(prds, sd.Admins[0].Products...)
	prds = append(prds, sd.Users[0].Products...)

	sort.Slice(prds, func(i, j int) bool {
		return prds[i].ID.String() <= prds[j].ID.String()
	})

	table := []table{
		{
			Name:       "basic",
			URL:        "/v1/products?page=1&rows=10&orderBy=product_id,ASC",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			Resp:       &page.Document[productapp.Product]{},
			ExpResp: &page.Document[productapp.Product]{
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

func productQueryByID200(sd seedData) []table {
	table := []table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/products/%s", sd.Users[0].Products[0].ID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			Resp:       &productapp.Product{},
			ExpResp:    toAppProductPtr(sd.Users[0].Products[0]),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
