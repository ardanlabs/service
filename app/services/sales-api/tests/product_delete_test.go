package tests

import (
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/web/errs"
	"github.com/google/go-cmp/cmp"
)

func productDelete200(sd dbtest.SeedData) []dbtest.AppTable {
	table := []dbtest.AppTable{
		{
			Name:       "asuser",
			URL:        fmt.Sprintf("/v1/products/%s", sd.Users[0].Products[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusNoContent,
		},
		{
			Name:       "asadmin",
			URL:        fmt.Sprintf("/v1/products/%s", sd.Admins[0].Products[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusNoContent,
		},
	}

	return table
}

func productDelete401(sd dbtest.SeedData) []dbtest.AppTable {
	table := []dbtest.AppTable{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/products/%s", sd.Users[0].Products[1].ID),
			Token:      "",
			Method:     http.MethodDelete,
			StatusCode: http.StatusUnauthorized,
			Resp:       &errs.Response{},
			ExpResp:    &errs.Response{Error: "Unauthorized"},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badsig",
			URL:        fmt.Sprintf("/v1/products/%s", sd.Users[0].Products[1].ID),
			Token:      sd.Users[0].Token + "A",
			Method:     http.MethodDelete,
			StatusCode: http.StatusUnauthorized,
			Resp:       &errs.Response{},
			ExpResp:    &errs.Response{Error: "Unauthorized"},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "wronguser",
			URL:        fmt.Sprintf("/v1/products/%s", sd.Admins[0].Products[1].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusUnauthorized,
			Resp:       &errs.Response{},
			ExpResp:    &errs.Response{Error: "Unauthorized"},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
