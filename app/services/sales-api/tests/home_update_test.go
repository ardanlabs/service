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

func homeUpdate200(sd dbtest.SeedData) []dbtest.AppTable {
	table := []dbtest.AppTable{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/homes/%s", sd.Users[0].Homes[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Model: &homegrp.AppUpdateHome{
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
			Resp: &homegrp.AppHome{},
			ExpResp: &homegrp.AppHome{
				ID:     sd.Users[0].Homes[0].ID.String(),
				UserID: sd.Users[0].ID.String(),
				Type:   "SINGLE FAMILY",
				Address: homegrp.AppAddress{
					Address1: "123 Mocking Bird Lane",
					Address2: "apt 105",
					ZipCode:  "35810",
					City:     "Huntsville",
					State:    "AL",
					Country:  "US",
				},
				DateCreated: sd.Users[0].Homes[0].DateCreated.Format(time.RFC3339),
				DateUpdated: sd.Users[0].Homes[0].DateCreated.Format(time.RFC3339),
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*homegrp.AppHome)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*homegrp.AppHome)
				gotResp.DateUpdated = expResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func homeUpdate400(sd dbtest.SeedData) []dbtest.AppTable {
	table := []dbtest.AppTable{
		{
			Name:       "bad-input",
			URL:        fmt.Sprintf("/v1/homes/%s", sd.Users[0].Homes[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Model: &homegrp.AppUpdateHome{
				Address: &homegrp.AppUpdateAddress{
					Address1: dbtest.StringPointer(""),
					Address2: dbtest.StringPointer(""),
					ZipCode:  dbtest.StringPointer(""),
					City:     dbtest.StringPointer(""),
					State:    dbtest.StringPointer(""),
					Country:  dbtest.StringPointer(""),
				},
			},
			Resp: &errs.Response{},
			ExpResp: &errs.Response{
				Error:  "data validation error",
				Fields: map[string]string{"address1": "address1 must be at least 1 character in length", "country": "Key: 'AppUpdateHome.address.country' Error:Field validation for 'country' failed on the 'iso3166_1_alpha2' tag", "state": "state must be at least 1 character in length", "zipCode": "zipCode must be a valid numeric value"},
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-type",
			URL:        fmt.Sprintf("/v1/homes/%s", sd.Users[0].Homes[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Model: &homegrp.AppUpdateHome{
				Type:    dbtest.StringPointer("BAD TYPE"),
				Address: &homegrp.AppUpdateAddress{},
			},
			Resp: &errs.Response{},
			ExpResp: &errs.Response{
				Error: "parse: invalid type \"BAD TYPE\"",
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func homeUpdate401(sd dbtest.SeedData) []dbtest.AppTable {
	table := []dbtest.AppTable{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/homes/%s", sd.Users[0].Homes[0].ID),
			Token:      "",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Resp:       &errs.Response{},
			ExpResp:    &errs.Response{Error: "Unauthorized"},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badsig",
			URL:        fmt.Sprintf("/v1/homes/%s", sd.Users[0].Homes[0].ID),
			Token:      sd.Users[0].Token + "A",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Resp:       &errs.Response{},
			ExpResp:    &errs.Response{Error: "Unauthorized"},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "wronguser",
			URL:        fmt.Sprintf("/v1/homes/%s", sd.Admins[0].Homes[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Model: &homegrp.AppUpdateHome{
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
			Resp:    &errs.Response{},
			ExpResp: &errs.Response{Error: "Unauthorized"},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
