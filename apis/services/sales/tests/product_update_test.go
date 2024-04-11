package tests

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ardanlabs/service/app/api/apptest"
	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/core/crud/productapp"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/google/go-cmp/cmp"
)

func productUpdate200(sd dbtest.SeedData) []apptest.AppTable {
	table := []apptest.AppTable{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/products/%s", sd.Users[0].Products[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Model: &productapp.UpdateProduct{
				Name:     dbtest.StringPointer("Guitar"),
				Cost:     dbtest.FloatPointer(10.34),
				Quantity: dbtest.IntPointer(10),
			},
			Resp: &productapp.Product{},
			ExpResp: &productapp.Product{
				ID:          sd.Users[0].Products[0].ID.String(),
				UserID:      sd.Users[0].ID.String(),
				Name:        "Guitar",
				Cost:        10.34,
				Quantity:    10,
				DateCreated: sd.Users[0].Products[0].DateCreated.Format(time.RFC3339),
				DateUpdated: sd.Users[0].Products[0].DateCreated.Format(time.RFC3339),
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*productapp.Product)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*productapp.Product)
				gotResp.DateUpdated = expResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func productUpdate400(sd dbtest.SeedData) []apptest.AppTable {
	table := []apptest.AppTable{
		{
			Name:       "bad-input",
			URL:        fmt.Sprintf("/v1/products/%s", sd.Users[0].Products[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Model: &productapp.UpdateProduct{
				Cost:     dbtest.FloatPointer(-1.0),
				Quantity: dbtest.IntPointer(0),
			},
			Resp:    &errs.Error{},
			ExpResp: toErrorPtr(errs.Newf(errs.FailedPrecondition, "validate: [{\"field\":\"cost\",\"error\":\"cost must be 0 or greater\"},{\"field\":\"quantity\",\"error\":\"quantity must be 1 or greater\"}]")),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func productUpdate401(sd dbtest.SeedData) []apptest.AppTable {
	table := []apptest.AppTable{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/products/%s", sd.Users[0].Products[0].ID),
			Token:      "",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Resp:       &errs.Error{},
			ExpResp:    toErrorPtr(errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments")),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badsig",
			URL:        fmt.Sprintf("/v1/products/%s", sd.Users[0].Products[0].ID),
			Token:      sd.Users[0].Token + "A",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Resp:       &errs.Error{},
			ExpResp:    toErrorPtr(errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]")),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "wronguser",
			URL:        fmt.Sprintf("/v1/products/%s", sd.Admins[0].Products[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Model: &productapp.UpdateProduct{
				Name:     dbtest.StringPointer("Guitar"),
				Cost:     dbtest.FloatPointer(10.34),
				Quantity: dbtest.IntPointer(10),
			},
			Resp:    &errs.Error{},
			ExpResp: toErrorPtr(errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[{USER}]] rule[rule_admin_or_subject]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]")),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
