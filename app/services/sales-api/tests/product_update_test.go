package tests

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/crud/productgrp"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/web/errs"
	"github.com/google/go-cmp/cmp"
)

func productUpdate200(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "basic",
			url:        fmt.Sprintf("/v1/products/%s", sd.users[0].products[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodPut,
			statusCode: http.StatusOK,
			model: &productgrp.AppUpdateProduct{
				Name:     dbtest.StringPointer("Guitar"),
				Cost:     dbtest.FloatPointer(10.34),
				Quantity: dbtest.IntPointer(10),
			},
			resp: &productgrp.AppProduct{},
			expResp: &productgrp.AppProduct{
				ID:          sd.users[0].products[0].ID.String(),
				UserID:      sd.users[0].ID.String(),
				Name:        "Guitar",
				Cost:        10.34,
				Quantity:    10,
				DateCreated: sd.users[0].products[0].DateCreated.Format(time.RFC3339),
				DateUpdated: sd.users[0].products[0].DateCreated.Format(time.RFC3339),
			},
			cmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*productgrp.AppProduct)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*productgrp.AppProduct)
				expResp.DateUpdated = gotResp.DateCreated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func productUpdate400(sd seedData) []tableData {
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
			resp: &errs.Response{},
			expResp: &errs.Response{
				Error:  "data validation error",
				Fields: map[string]string{"cost": "cost must be 0 or greater", "quantity": "quantity must be 1 or greater"},
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func productUpdate401(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "emptytoken",
			url:        fmt.Sprintf("/v1/products/%s", sd.users[0].products[0].ID),
			token:      "",
			method:     http.MethodPut,
			statusCode: http.StatusUnauthorized,
			resp:       &errs.Response{},
			expResp:    &errs.Response{Error: "Unauthorized"},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			name:       "badsig",
			url:        fmt.Sprintf("/v1/products/%s", sd.users[0].products[0].ID),
			token:      sd.users[0].token + "A",
			method:     http.MethodPut,
			statusCode: http.StatusUnauthorized,
			resp:       &errs.Response{},
			expResp:    &errs.Response{Error: "Unauthorized"},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			name:       "wronguser",
			url:        fmt.Sprintf("/v1/products/%s", sd.admins[0].products[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodPut,
			statusCode: http.StatusUnauthorized,
			model: &productgrp.AppUpdateProduct{
				Name:     dbtest.StringPointer("Guitar"),
				Cost:     dbtest.FloatPointer(10.34),
				Quantity: dbtest.IntPointer(10),
			},
			resp:    &errs.Response{},
			expResp: &errs.Response{Error: "Unauthorized"},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
