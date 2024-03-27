package tests

import (
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/business/web/errs"
	"github.com/google/go-cmp/cmp"
)

func productDelete200(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "asuser",
			url:        fmt.Sprintf("/v1/products/%s", sd.users[0].products[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "asadmin",
			url:        fmt.Sprintf("/v1/products/%s", sd.admins[0].products[0].ID),
			token:      sd.admins[0].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
	}

	return table
}

func productDelete401(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "emptytoken",
			url:        fmt.Sprintf("/v1/products/%s", sd.users[0].products[1].ID),
			token:      "",
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &errs.Response{},
			expResp:    &errs.Response{Error: "Unauthorized"},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			name:       "badsig",
			url:        fmt.Sprintf("/v1/products/%s", sd.users[0].products[1].ID),
			token:      sd.users[0].token + "A",
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &errs.Response{},
			expResp:    &errs.Response{Error: "Unauthorized"},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			name:       "wronguser",
			url:        fmt.Sprintf("/v1/products/%s", sd.admins[0].products[1].ID),
			token:      sd.users[0].token,
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &errs.Response{},
			expResp:    &errs.Response{Error: "Unauthorized"},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
