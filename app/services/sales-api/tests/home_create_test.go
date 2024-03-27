package tests

import (
	"net/http"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/crud/homegrp"
	"github.com/ardanlabs/service/business/web/errs"
	"github.com/google/go-cmp/cmp"
)

func homeCreate200(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "basic",
			url:        "/v1/homes",
			token:      sd.users[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusCreated,
			model: &homegrp.AppNewHome{
				Type: "SINGLE FAMILY",
				Address: homegrp.AppNewAddress{
					Address1: "123 Mocking Bird Lane",
					ZipCode:  "35810",
					City:     "Huntsville",
					State:    "AL",
					Country:  "US",
				},
			},
			resp: &homegrp.AppHome{},
			expResp: &homegrp.AppHome{
				UserID: sd.users[0].ID.String(),
				Type:   "SINGLE FAMILY",
				Address: homegrp.AppAddress{
					Address1: "123 Mocking Bird Lane",
					ZipCode:  "35810",
					City:     "Huntsville",
					State:    "AL",
					Country:  "US",
				},
			},
			cmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*homegrp.AppHome)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*homegrp.AppHome)

				expResp.ID = gotResp.ID
				expResp.DateCreated = gotResp.DateCreated
				expResp.DateUpdated = gotResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func homeCreate400(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "missing-input",
			url:        "/v1/homes",
			token:      sd.users[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusBadRequest,
			model:      &homegrp.AppNewHome{},
			resp:       &errs.Response{},
			expResp: &errs.Response{
				Error:  "data validation error",
				Fields: map[string]string{"address1": "address1 is a required field", "city": "city is a required field", "country": "country is a required field", "state": "state is a required field", "type": "type is a required field", "zipCode": "zipCode is a required field"},
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
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
			resp: &errs.Response{},
			expResp: &errs.Response{
				Error: "parse: invalid type \"BAD TYPE\"",
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func homeCreate401(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "emptytoken",
			url:        "/v1/homes",
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
			url:        "/v1/homes",
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
			url:        "/v1/homes",
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
			url:        "/v1/homes",
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
