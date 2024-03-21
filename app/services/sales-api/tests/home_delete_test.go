package tests

import (
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/business/web/errs"
	"github.com/google/go-cmp/cmp"
)

func homeDelete200(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "asuser",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "asadmin",
			url:        fmt.Sprintf("/v1/homes/%s", sd.admins[0].homes[0].ID),
			token:      sd.admins[0].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
	}

	return table
}

func homeDelete401(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "emptytoken",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[1].ID),
			token:      "",
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &errs.Response{},
			expResp:    &errs.Response{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "badsig",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[1].ID),
			token:      sd.users[0].token + "A",
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &errs.Response{},
			expResp:    &errs.Response{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "wronguser",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[1].ID),
			token:      sd.users[1].token,
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &errs.Response{},
			expResp:    &errs.Response{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}
