package tests

import (
	"net/http"
	"testing"

	"github.com/ardanlabs/service/business/web/v1/response"
	"github.com/google/go-cmp/cmp"
)

func userCreate401(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "emptytoken",
			url:        "/v1/users",
			token:      "",
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "badtoken",
			url:        "/v1/users",
			token:      sd.admins[0].token[:10],
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "badsig",
			url:        "/v1/users",
			token:      sd.admins[0].token + "A",
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "wronguser",
			url:        "/v1/users",
			token:      sd.users[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}

func productCreate401(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "emptytoken",
			url:        "/v1/products",
			token:      "",
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "badtoken",
			url:        "/v1/products",
			token:      sd.admins[0].token[:10],
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "badsig",
			url:        "/v1/products",
			token:      sd.admins[0].token + "A",
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "wronguser",
			url:        "/v1/products",
			token:      sd.admins[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}

func homeCreate401(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "emptytoken",
			url:        "/v1/homes",
			token:      "",
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "badtoken",
			url:        "/v1/homes",
			token:      sd.admins[0].token[:10],
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "badsig",
			url:        "/v1/homes",
			token:      sd.admins[0].token + "A",
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "wronguser",
			url:        "/v1/homes",
			token:      sd.admins[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}
