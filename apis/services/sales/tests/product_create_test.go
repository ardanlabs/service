package tests

import (
	"net/http"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/domain/productapp"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/google/go-cmp/cmp"
)

func productCreate200(sd dbtest.SeedData) []appTable {
	table := []appTable{
		{
			Name:       "basic",
			URL:        "/v1/products",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusCreated,
			Model: &productapp.NewProduct{
				Name:     "Guitar",
				Cost:     10.34,
				Quantity: 10,
			},
			Resp: &productapp.Product{},
			ExpResp: &productapp.Product{
				Name:     "Guitar",
				UserID:   sd.Users[0].ID.String(),
				Cost:     10.34,
				Quantity: 10,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*productapp.Product)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*productapp.Product)

				expResp.ID = gotResp.ID
				expResp.DateCreated = gotResp.DateCreated
				expResp.DateUpdated = gotResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func productCreate400(sd dbtest.SeedData) []appTable {
	table := []appTable{
		{
			Name:       "missing-input",
			URL:        "/v1/products",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Model:      &productapp.NewProduct{},
			Resp:       &errs.Error{},
			ExpResp:    toErrorPtr(errs.Newf(errs.FailedPrecondition, "validate: [{\"field\":\"name\",\"error\":\"name is a required field\"},{\"field\":\"cost\",\"error\":\"cost is a required field\"},{\"field\":\"quantity\",\"error\":\"quantity is a required field\"}]")),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func productCreate401(sd dbtest.SeedData) []appTable {
	table := []appTable{
		{
			Name:       "emptytoken",
			URL:        "/v1/products",
			Token:      "&nbsp;",
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
			URL:        "/v1/products",
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
			URL:        "/v1/products",
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
			URL:        "/v1/products",
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
