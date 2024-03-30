package tests

import (
	"net/http"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/crud/productgrp"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/web/errs"
	"github.com/google/go-cmp/cmp"
)

func productCreate200(sd dbtest.SeedData) []dbtest.AppTable {
	table := []dbtest.AppTable{
		{
			Name:       "basic",
			URL:        "/v1/products",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusCreated,
			Model: &productgrp.AppNewProduct{
				Name:     "Guitar",
				Cost:     10.34,
				Quantity: 10,
			},
			Resp: &productgrp.AppProduct{},
			ExpResp: &productgrp.AppProduct{
				Name:     "Guitar",
				UserID:   sd.Users[0].ID.String(),
				Cost:     10.34,
				Quantity: 10,
			},
			CmpFunc: func(got any, exp any) string {
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

func productCreate400(sd dbtest.SeedData) []dbtest.AppTable {
	table := []dbtest.AppTable{
		{
			Name:       "missing-input",
			URL:        "/v1/products",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Model:      &productgrp.AppNewProduct{},
			Resp:       &errs.Response{},
			ExpResp: &errs.Response{
				Error:  "data validation error",
				Fields: map[string]string{"cost": "cost is a required field", "name": "name is a required field", "quantity": "quantity is a required field"},
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func productCreate401(sd dbtest.SeedData) []dbtest.AppTable {
	table := []dbtest.AppTable{
		{
			Name:       "emptytoken",
			URL:        "/v1/products",
			Token:      "",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Resp:       &errs.Response{},
			ExpResp: &errs.Response{
				Error: "Unauthorized",
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badtoken",
			URL:        "/v1/products",
			Token:      sd.Admins[0].Token[:10],
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Resp:       &errs.Response{},
			ExpResp: &errs.Response{
				Error: "Unauthorized",
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badsig",
			URL:        "/v1/products",
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Resp:       &errs.Response{},
			ExpResp: &errs.Response{
				Error: "Unauthorized",
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "wronguser",
			URL:        "/v1/products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Resp:       &errs.Response{},
			ExpResp: &errs.Response{
				Error: "Unauthorized",
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
