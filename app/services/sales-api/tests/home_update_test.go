package tests

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/crud/homegrp"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/web/errs"
	"github.com/google/go-cmp/cmp"
)

func homeUpdate200(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "basic",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodPut,
			statusCode: http.StatusOK,
			model: &homegrp.AppUpdateHome{
				Type: dbtest.StringPointer("SINGLE FAMILY"),
				Address: &homegrp.AppUpdateAddress{
					Address1: dbtest.StringPointer("123 Mocking Bird Lane"),
					Address2: dbtest.StringPointer("apt 105"),
					ZipCode:  dbtest.StringPointer("35810"),
					City:     dbtest.StringPointer("Huntsville"),
					State:    dbtest.StringPointer("AL"),
					Country:  dbtest.StringPointer("US"),
				},
			},
			resp: &homegrp.AppHome{},
			expResp: &homegrp.AppHome{
				ID:     sd.users[0].homes[0].ID.String(),
				UserID: sd.users[0].ID.String(),
				Type:   "SINGLE FAMILY",
				Address: homegrp.AppAddress{
					Address1: "123 Mocking Bird Lane",
					Address2: "apt 105",
					ZipCode:  "35810",
					City:     "Huntsville",
					State:    "AL",
					Country:  "US",
				},
				DateCreated: sd.users[0].homes[0].DateCreated.Format(time.RFC3339),
				DateUpdated: sd.users[0].homes[0].DateCreated.Format(time.RFC3339),
			},
			cmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*homegrp.AppHome)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*homegrp.AppHome)
				expResp.DateUpdated = gotResp.DateCreated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func homeUpdate400(sd seedData) []tableData {
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
			resp: &errs.Response{},
			expResp: &errs.Response{
				Error:  "data validation error",
				Fields: map[string]string{"address1": "address1 must be at least 1 character in length", "country": "Key: 'AppUpdateHome.address.country' Error:Field validation for 'country' failed on the 'iso3166_1_alpha2' tag", "state": "state must be at least 1 character in length", "zipCode": "zipCode must be a valid numeric value"},
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
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

func homeUpdate401(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "emptytoken",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[0].ID),
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
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[0].ID),
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
			url:        fmt.Sprintf("/v1/homes/%s", sd.admins[0].homes[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodPut,
			statusCode: http.StatusUnauthorized,
			model: &homegrp.AppUpdateHome{
				Type: dbtest.StringPointer("SINGLE FAMILY"),
				Address: &homegrp.AppUpdateAddress{
					Address1: dbtest.StringPointer("123 Mocking Bird Lane"),
					Address2: dbtest.StringPointer("apt 105"),
					ZipCode:  dbtest.StringPointer("35810"),
					City:     dbtest.StringPointer("Huntsville"),
					State:    dbtest.StringPointer("AL"),
					Country:  dbtest.StringPointer("US"),
				},
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
