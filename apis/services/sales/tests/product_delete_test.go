package tests

import (
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/app/api/apptest"
	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/google/go-cmp/cmp"
)

func productDelete200(sd dbtest.SeedData) []apptest.AppTable {
	table := []apptest.AppTable{
		{
			Name:       "asuser",
			URL:        fmt.Sprintf("/v1/products/%s", sd.Users[0].Products[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusNoContent,
		},
		{
			Name:       "asadmin",
			URL:        fmt.Sprintf("/v1/products/%s", sd.Admins[0].Products[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusNoContent,
		},
	}

	return table
}

func productDelete401(sd dbtest.SeedData) []apptest.AppTable {
	table := []apptest.AppTable{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/products/%s", sd.Users[0].Products[1].ID),
			Token:      "&nbsp;",
			Method:     http.MethodDelete,
			StatusCode: http.StatusUnauthorized,
			Resp:       &errs.Error{},
			ExpResp:    toErrorPtr(errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments")),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badsig",
			URL:        fmt.Sprintf("/v1/products/%s", sd.Users[0].Products[1].ID),
			Token:      sd.Users[0].Token + "A",
			Method:     http.MethodDelete,
			StatusCode: http.StatusUnauthorized,
			Resp:       &errs.Error{},
			ExpResp:    toErrorPtr(errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]")),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "wronguser",
			URL:        fmt.Sprintf("/v1/products/%s", sd.Admins[0].Products[1].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusUnauthorized,
			Resp:       &errs.Error{},
			ExpResp:    toErrorPtr(errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[{USER}]] rule[rule_admin_or_subject]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]")),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
