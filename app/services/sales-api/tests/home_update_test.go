package tests

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ardanlabs/service/app/services/sales-api/apis/crud/homeapi"
	"github.com/ardanlabs/service/business/api/errs"
	"github.com/ardanlabs/service/business/data/dbtest"
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
			Model: &homeapi.AppUpdateHome{
				Type: dbtest.StringPointer("SINGLE FAMILY"),
				Address: &homeapi.AppUpdateAddress{
					Address1: dbtest.StringPointer("123 Mocking Bird Lane"),
					Address2: dbtest.StringPointer("apt 105"),
					ZipCode:  dbtest.StringPointer("35810"),
					City:     dbtest.StringPointer("Huntsville"),
					State:    dbtest.StringPointer("AL"),
					Country:  dbtest.StringPointer("US"),
				},
			},
			Resp: &homeapi.AppHome{},
			ExpResp: &homeapi.AppHome{
				ID:     sd.Users[0].Homes[0].ID.String(),
				UserID: sd.Users[0].ID.String(),
				Type:   "SINGLE FAMILY",
				Address: homeapi.AppAddress{
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
				gotResp, exists := got.(*homeapi.AppHome)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*homeapi.AppHome)
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
			Model: &homeapi.AppUpdateHome{
				Address: &homeapi.AppUpdateAddress{
					Address1: dbtest.StringPointer(""),
					Address2: dbtest.StringPointer(""),
					ZipCode:  dbtest.StringPointer(""),
					City:     dbtest.StringPointer(""),
					State:    dbtest.StringPointer(""),
					Country:  dbtest.StringPointer(""),
				},
			},
			Resp:    &errs.Error{},
			ExpResp: toErrorPtr(errs.Newf(http.StatusBadRequest, "validate: [{\"field\":\"address1\",\"error\":\"address1 must be at least 1 character in length\"},{\"field\":\"zipCode\",\"error\":\"zipCode must be a valid numeric value\"},{\"field\":\"state\",\"error\":\"state must be at least 1 character in length\"},{\"field\":\"country\",\"error\":\"Key: 'AppUpdateHome.address.country' Error:Field validation for 'country' failed on the 'iso3166_1_alpha2' tag\"}]")),
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
			Model: &homeapi.AppUpdateHome{
				Type:    dbtest.StringPointer("BAD TYPE"),
				Address: &homeapi.AppUpdateAddress{},
			},
			Resp:    &errs.Error{},
			ExpResp: toErrorPtr(errs.Newf(http.StatusBadRequest, "parse: invalid type \"BAD TYPE\"")),
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
			Resp:       &errs.Error{},
			ExpResp:    toErrorPtr(errs.Newf(http.StatusUnauthorized, "error parsing token: token contains an invalid number of segments")),
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
			Resp:       &errs.Error{},
			ExpResp:    toErrorPtr(errs.Newf(http.StatusUnauthorized, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]")),
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
			Model: &homeapi.AppUpdateHome{
				Type: dbtest.StringPointer("SINGLE FAMILY"),
				Address: &homeapi.AppUpdateAddress{
					Address1: dbtest.StringPointer("123 Mocking Bird Lane"),
					Address2: dbtest.StringPointer("apt 105"),
					ZipCode:  dbtest.StringPointer("35810"),
					City:     dbtest.StringPointer("Huntsville"),
					State:    dbtest.StringPointer("AL"),
					Country:  dbtest.StringPointer("US"),
				},
			},
			Resp:    &errs.Error{},
			ExpResp: toErrorPtr(errs.Newf(http.StatusUnauthorized, "authorize: you are not authorized for that action, claims[[{USER}]] rule[rule_admin_or_subject]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]")),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
