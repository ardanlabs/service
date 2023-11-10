package tests

import (
	"net/http"
	"testing"

	"github.com/ardanlabs/service/business/web/v1/response"
	"github.com/google/go-cmp/cmp"
)

func testCreate401(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "user-emptytoken",
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
			name:       "user-badtoken",
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
			name:       "user-badsig",
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
			name:       "user-wronguser",
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
		{
			name:       "product-emptytoken",
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
			name:       "product-badtoken",
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
			name:       "product-badsig",
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
			name:       "product-wronguser",
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
		{
			name:       "home-emptytoken",
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
			name:       "home-badtoken",
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
			name:       "home-badsig",
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
			name:       "home-wronguser",
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
