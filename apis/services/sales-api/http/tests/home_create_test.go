package tests

import (
	"net/http"

	"github.com/ardanlabs/service/app/core/crud/homeapp"
	"github.com/ardanlabs/service/business/api/errs"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/google/go-cmp/cmp"
)

func homeCreate200(sd dbtest.SeedData) []dbtest.AppTable {
	table := []dbtest.AppTable{
		{
			Name:       "basic",
			URL:        "/v1/homes",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusCreated,
			Model: &homeapp.NewHome{
				Type: "SINGLE FAMILY",
				Address: homeapp.NewAddress{
					Address1: "123 Mocking Bird Lane",
					ZipCode:  "35810",
					City:     "Huntsville",
					State:    "AL",
					Country:  "US",
				},
			},
			Resp: &homeapp.Home{},
			ExpResp: &homeapp.Home{
				UserID: sd.Users[0].ID.String(),
				Type:   "SINGLE FAMILY",
				Address: homeapp.Address{
					Address1: "123 Mocking Bird Lane",
					ZipCode:  "35810",
					City:     "Huntsville",
					State:    "AL",
					Country:  "US",
				},
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*homeapp.Home)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*homeapp.Home)

				expResp.ID = gotResp.ID
				expResp.DateCreated = gotResp.DateCreated
				expResp.DateUpdated = gotResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func homeCreate400(sd dbtest.SeedData) []dbtest.AppTable {
	table := []dbtest.AppTable{
		{
			Name:       "missing-input",
			URL:        "/v1/homes",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Model:      &homeapp.NewHome{},
			Resp:       &errs.Error{},
			ExpResp:    toErrorPtr(errs.Newf(errs.FailedPrecondition, "validate: [{\"field\":\"type\",\"error\":\"type is a required field\"},{\"field\":\"address1\",\"error\":\"address1 is a required field\"},{\"field\":\"zipCode\",\"error\":\"zipCode is a required field\"},{\"field\":\"city\",\"error\":\"city is a required field\"},{\"field\":\"state\",\"error\":\"state is a required field\"},{\"field\":\"country\",\"error\":\"country is a required field\"}]")),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-type",
			URL:        "/v1/homes",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Model: &homeapp.NewHome{
				Type: "BAD TYPE",
				Address: homeapp.NewAddress{
					Address1: "123 Mocking Bird Lane",
					ZipCode:  "35810",
					City:     "Huntsville",
					State:    "AL",
					Country:  "US",
				},
			},
			Resp:    &errs.Error{},
			ExpResp: toErrorPtr(errs.Newf(errs.FailedPrecondition, "parse: invalid type \"BAD TYPE\"")),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func homeCreate401(sd dbtest.SeedData) []dbtest.AppTable {
	table := []dbtest.AppTable{
		{
			Name:       "emptytoken",
			URL:        "/v1/homes",
			Token:      "",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Resp:       &errs.Error{},
			ExpResp:    toErrorPtr(errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments")),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badtoken",
			URL:        "/v1/homes",
			Token:      sd.Admins[0].Token[:10],
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Resp:       &errs.Error{},
			ExpResp:    toErrorPtr(errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments")),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badsig",
			URL:        "/v1/homes",
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Resp:       &errs.Error{},
			ExpResp:    toErrorPtr(errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]")),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "wronguser",
			URL:        "/v1/homes",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Resp:       &errs.Error{},
			ExpResp:    toErrorPtr(errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[{ADMIN}]] rule[rule_user_only]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]")),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
