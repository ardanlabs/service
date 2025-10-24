package home_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ardanlabs/service/app/domain/homeapp"
	"github.com/ardanlabs/service/app/sdk/apitest"
	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/business/sdk/dbtest"
	"github.com/google/go-cmp/cmp"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/homes/%s", sd.Users[0].Homes[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &homeapp.UpdateHome{
				Type: dbtest.StringPointer("SINGLE FAMILY"),
				Address: &homeapp.UpdateAddress{
					Address1: dbtest.StringPointer("123 Mocking Bird Lane"),
					Address2: dbtest.StringPointer("apt 105"),
					ZipCode:  dbtest.StringPointer("35810"),
					City:     dbtest.StringPointer("Huntsville"),
					State:    dbtest.StringPointer("AL"),
					Country:  dbtest.StringPointer("US"),
				},
			},
			GotResp: &homeapp.Home{},
			ExpResp: &homeapp.Home{
				ID:     sd.Users[0].Homes[0].ID.String(),
				UserID: sd.Users[0].ID.String(),
				Type:   "SINGLE FAMILY",
				Address: homeapp.Address{
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
				gotResp, exists := got.(*homeapp.Home)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*homeapp.Home)
				gotResp.DateUpdated = expResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "bad-type",
			URL:        fmt.Sprintf("/v1/homes/%s", sd.Users[0].Homes[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &homeapp.UpdateHome{
				Type:    dbtest.StringPointer("BAD TYPE"),
				Address: &homeapp.UpdateAddress{},
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"type\",\"error\":\"invalid home type \\\"BAD TYPE\\\"\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func update401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/homes/%s", sd.Users[0].Homes[0].ID),
			Token:      "&nbsp;",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
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
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed: OPA policy evaluation failed for authentication: OPA policy rule \"auth\" not satisfied"),
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
			Input: &homeapp.UpdateHome{
				Type: dbtest.StringPointer("SINGLE FAMILY"),
				Address: &homeapp.UpdateAddress{
					Address1: dbtest.StringPointer("123 Mocking Bird Lane"),
					Address2: dbtest.StringPointer("apt 105"),
					ZipCode:  dbtest.StringPointer("35810"),
					City:     dbtest.StringPointer("Huntsville"),
					State:    dbtest.StringPointer("AL"),
					Country:  dbtest.StringPointer("US"),
				},
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_or_subject]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
