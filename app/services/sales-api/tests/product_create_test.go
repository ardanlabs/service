package tests

import (
	"net/http"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/crud/productgrp"
	"github.com/ardanlabs/service/business/web/errs"
	"github.com/google/go-cmp/cmp"
)

func productCreate200(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "basic",
			url:        "/v1/products",
			token:      sd.users[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusCreated,
			model: &productgrp.AppNewProduct{
				Name:     "Guitar",
				Cost:     10.34,
				Quantity: 10,
			},
			resp: &productgrp.AppProduct{},
			expResp: &productgrp.AppProduct{
				Name:     "Guitar",
				UserID:   sd.users[0].ID.String(),
				Cost:     10.34,
				Quantity: 10,
			},
			cmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*productgrp.AppProduct)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*productgrp.AppProduct)

				expResp.ID = gotResp.ID
				expResp.DateCreated = gotResp.DateCreated
				expResp.DateUpdated = gotResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func productCreate400(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "missing-input",
			url:        "/v1/products",
			token:      sd.users[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusBadRequest,
			model:      &productgrp.AppNewProduct{},
			resp:       &errs.Response{},
			expResp: &errs.Response{
				Error:  "data validation error",
				Fields: map[string]string{"cost": "cost is a required field", "name": "name is a required field", "quantity": "quantity is a required field"},
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func productCreate401(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "emptytoken",
			url:        "/v1/products",
			token:      "",
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &errs.Response{},
			expResp: &errs.Response{
				Error: "Unauthorized",
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			name:       "badtoken",
			url:        "/v1/products",
			token:      sd.admins[0].token[:10],
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &errs.Response{},
			expResp: &errs.Response{
				Error: "Unauthorized",
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			name:       "badsig",
			url:        "/v1/products",
			token:      sd.admins[0].token + "A",
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &errs.Response{},
			expResp: &errs.Response{
				Error: "Unauthorized",
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			name:       "wronguser",
			url:        "/v1/products",
			token:      sd.admins[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &errs.Response{},
			expResp: &errs.Response{
				Error: "Unauthorized",
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
