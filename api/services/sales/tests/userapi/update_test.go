package user_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ardanlabs/service/app/domain/userapp"
	"github.com/ardanlabs/service/app/sdk/apitest"
	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/business/sdk/dbtest"
	"github.com/google/go-cmp/cmp"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/users/%s", sd.Users[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &userapp.UpdateUser{
				Name:            dbtest.StringPointer("Jack Kennedy"),
				Email:           dbtest.StringPointer("jack@ardanlabs.com"),
				Department:      dbtest.StringPointer("ITO"),
				Password:        dbtest.StringPointer("123"),
				PasswordConfirm: dbtest.StringPointer("123"),
			},
			GotResp: &userapp.User{},
			ExpResp: &userapp.User{
				ID:          sd.Users[0].ID.String(),
				Name:        "Jack Kennedy",
				Email:       "jack@ardanlabs.com",
				Roles:       []string{"USER"},
				Department:  "ITO",
				Enabled:     true,
				DateCreated: sd.Users[0].DateCreated.Format(time.RFC3339),
				DateUpdated: sd.Users[0].DateUpdated.Format(time.RFC3339),
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*userapp.User)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*userapp.User)
				gotResp.DateUpdated = expResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:       "role",
			URL:        fmt.Sprintf("/v1/users/role/%s", sd.Admins[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &userapp.UpdateUserRole{
				Roles: []string{"USER"},
			},
			GotResp: &userapp.User{},
			ExpResp: &userapp.User{
				ID:          sd.Admins[0].ID.String(),
				Name:        sd.Admins[0].Name.String(),
				Email:       sd.Admins[0].Email.Address,
				Roles:       []string{"USER"},
				Department:  sd.Admins[0].Department.String(),
				Enabled:     true,
				DateCreated: sd.Admins[0].DateCreated.Format(time.RFC3339),
				DateUpdated: sd.Admins[0].DateUpdated.Format(time.RFC3339),
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*userapp.User)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*userapp.User)
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
			Name:       "bad-input",
			URL:        fmt.Sprintf("/v1/users/%s", sd.Users[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &userapp.UpdateUser{
				Email:           dbtest.StringPointer("bill@"),
				PasswordConfirm: dbtest.StringPointer("jack"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"email\",\"error\":\"email must be a valid email address\"},{\"field\":\"passwordConfirm\",\"error\":\"passwordConfirm must be equal to Password\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-role",
			URL:        fmt.Sprintf("/v1/users/role/%s", sd.Admins[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &userapp.UpdateUserRole{
				Roles: []string{"BAD ROLE"},
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "parse: invalid role \"BAD ROLE\""),
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
			URL:        fmt.Sprintf("/v1/users/%s", sd.Users[0].ID),
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
			URL:        fmt.Sprintf("/v1/users/%s", sd.Users[0].ID),
			Token:      sd.Users[0].Token + "A",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "wronguser",
			URL:        fmt.Sprintf("/v1/users/%s", sd.Admins[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input: &userapp.UpdateUser{
				Name:            dbtest.StringPointer("Bill Kennedy"),
				Email:           dbtest.StringPointer("bill@ardanlabs.com"),
				Department:      dbtest.StringPointer("ITO"),
				Password:        dbtest.StringPointer("123"),
				PasswordConfirm: dbtest.StringPointer("123"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_or_subject]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "roleadminonly",
			URL:        fmt.Sprintf("/v1/users/role/%s", sd.Users[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input: &userapp.UpdateUserRole{
				Roles: []string{"ADMIN"},
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_only]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
