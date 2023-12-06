package tests

import (
	"fmt"
	"net/http"
	"testing"

	v1 "github.com/ardanlabs/service/business/web/v1"
	"github.com/google/go-cmp/cmp"
)

func userDelete200(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "asuser",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[1].ID),
			token:      sd.users[1].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "asadmin",
			url:        fmt.Sprintf("/v1/users/%s", sd.admins[1].ID),
			token:      sd.admins[1].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
	}

	return table
}

func userDelete401(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "emptytoken",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
			token:      "",
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &v1.ErrorResponse{},
			expResp:    &v1.ErrorResponse{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "badsig",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
			token:      sd.users[0].token + "A",
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &v1.ErrorResponse{},
			expResp:    &v1.ErrorResponse{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "wronguser",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
			token:      app.userToken,
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &v1.ErrorResponse{},
			expResp:    &v1.ErrorResponse{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}
