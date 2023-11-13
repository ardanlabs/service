package tests

import (
	"net/http"
	"testing"

	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/homegrp"
	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/productgrp"
	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/usergrp"
	"github.com/ardanlabs/service/business/web/v1/response"
	"github.com/google/go-cmp/cmp"
)

func userCreate400(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "missing-input",
			url:        "/v1/users",
			token:      sd.admins[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusBadRequest,
			model:      &usergrp.AppNewUser{},
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error:  "data validation error",
				Fields: map[string]string{"email": "email is a required field", "name": "name is a required field", "password": "password is a required field", "roles": "roles is a required field"},
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "bad-role",
			url:        "/v1/users",
			token:      sd.admins[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusBadRequest,
			model: &usergrp.AppNewUser{
				Name:            "Bill Kennedy",
				Email:           "bill@ardanlabs.com",
				Roles:           []string{"BAD ROLE"},
				Department:      "IT",
				Password:        "123",
				PasswordConfirm: "123",
			},
			resp: &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "parse: invalid role \"BAD ROLE\"",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}

func productCreate400(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "missing-input",
			url:        "/v1/products",
			token:      sd.users[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusBadRequest,
			model:      &productgrp.AppNewProduct{},
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error:  "data validation error",
				Fields: map[string]string{"cost": "cost is a required field", "name": "name is a required field", "quantity": "quantity must be 1 or greater"},
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}

func homeCreate400(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "missing-input",
			url:        "/v1/homes",
			token:      sd.users[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusBadRequest,
			model:      &homegrp.AppNewHome{},
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error:  "data validation error",
				Fields: map[string]string{"address1": "address1 is a required field", "country": "country is a required field", "state": "state is a required field", "type": "type is a required field"},
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "bad-type",
			url:        "/v1/homes",
			token:      sd.users[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusBadRequest,
			model: &homegrp.AppNewHome{
				Type: "BAD TYPE",
				Address: homegrp.AppNewAddress{
					Address1: "123 Mocking Bird Lane",
					ZipCode:  "35810",
					City:     "Huntsville",
					State:    "AL",
					Country:  "US",
				},
			},
			resp: &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "parse: invalid type \"BAD TYPE\"",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}
