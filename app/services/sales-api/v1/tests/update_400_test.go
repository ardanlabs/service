package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/homegrp"
	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/productgrp"
	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/usergrp"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/web/v1/response"
	"github.com/google/go-cmp/cmp"
)

func userUpdate400(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "bad-input",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodPut,
			statusCode: http.StatusBadRequest,
			model: &usergrp.AppUpdateUser{
				Email:           dbtest.StringPointer("bill@"),
				PasswordConfirm: dbtest.StringPointer("jack"),
			},
			resp: &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error:  "data validation error",
				Fields: map[string]string{"email": "email must be a valid email address", "passwordConfirm": "passwordConfirm must be equal to Password"},
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "bad-role",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodPut,
			statusCode: http.StatusBadRequest,
			model: &usergrp.AppUpdateUser{
				Roles: []string{"BAD ROLE"},
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

func productUpdate400(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "bad-input",
			url:        fmt.Sprintf("/v1/products/%s", sd.users[0].products[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodPut,
			statusCode: http.StatusBadRequest,
			model: &productgrp.AppUpdateProduct{
				Cost:     dbtest.FloatPointer(-1.0),
				Quantity: dbtest.IntPointer(0),
			},
			resp: &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error:  "data validation error",
				Fields: map[string]string{"cost": "cost must be 0 or greater", "quantity": "quantity must be 1 or greater"},
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}

func homeUpdate400(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "bad-input",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodPut,
			statusCode: http.StatusBadRequest,
			model: &homegrp.AppUpdateHome{
				Address: &homegrp.AppUpdateAddress{
					Address1: dbtest.StringPointer(""),
					Address2: dbtest.StringPointer(""),
					ZipCode:  dbtest.StringPointer(""),
					City:     dbtest.StringPointer(""),
					State:    dbtest.StringPointer(""),
					Country:  dbtest.StringPointer(""),
				},
			},
			resp: &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error:  "data validation error",
				Fields: map[string]string{"address1": "address1 must be at least 1 character in length", "country": "Key: 'AppUpdateHome.address.country' Error:Field validation for 'country' failed on the 'iso3166_1_alpha2' tag", "state": "state must be at least 1 character in length"},
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "bad-type",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodPut,
			statusCode: http.StatusBadRequest,
			model: &homegrp.AppUpdateHome{
				Type:    dbtest.StringPointer("BAD TYPE"),
				Address: &homegrp.AppUpdateAddress{},
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
