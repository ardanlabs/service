package tests

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/crud/usergrp"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/web/errs"
	"github.com/google/go-cmp/cmp"
)

func userUpdate200(sd dbtest.SeedData) []dbtest.AppTable {
	table := []dbtest.AppTable{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/users/%s", sd.Users[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Model: &usergrp.AppUpdateUser{
				Name:            dbtest.StringPointer("Jack Kennedy"),
				Email:           dbtest.StringPointer("jack@ardanlabs.com"),
				Roles:           []string{"ADMIN"},
				Department:      dbtest.StringPointer("IT"),
				Password:        dbtest.StringPointer("123"),
				PasswordConfirm: dbtest.StringPointer("123"),
			},
			Resp: &usergrp.AppUser{},
			ExpResp: &usergrp.AppUser{
				ID:          sd.Users[0].ID.String(),
				Name:        "Jack Kennedy",
				Email:       "jack@ardanlabs.com",
				Roles:       []string{"ADMIN"},
				Department:  "IT",
				Enabled:     true,
				DateCreated: sd.Users[0].DateCreated.Format(time.RFC3339),
				DateUpdated: sd.Users[0].DateUpdated.Format(time.RFC3339),
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*usergrp.AppUser)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*usergrp.AppUser)
				gotResp.DateUpdated = expResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func userUpdate400(sd dbtest.SeedData) []dbtest.AppTable {
	table := []dbtest.AppTable{
		{
			Name:       "bad-input",
			URL:        fmt.Sprintf("/v1/users/%s", sd.Users[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Model: &usergrp.AppUpdateUser{
				Email:           dbtest.StringPointer("bill@"),
				PasswordConfirm: dbtest.StringPointer("jack"),
			},
			Resp: &errs.Response{},
			ExpResp: &errs.Response{
				Error:  "data validation error",
				Fields: map[string]string{"email": "email must be a valid email address", "passwordConfirm": "passwordConfirm must be equal to Password"},
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-role",
			URL:        fmt.Sprintf("/v1/users/%s", sd.Users[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Model: &usergrp.AppUpdateUser{
				Roles: []string{"BAD ROLE"},
			},
			Resp: &errs.Response{},
			ExpResp: &errs.Response{
				Error: "parse: invalid role \"BAD ROLE\"",
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func userUpdate401(sd dbtest.SeedData) []dbtest.AppTable {
	table := []dbtest.AppTable{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/users/%s", sd.Users[0].ID),
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
			URL:        fmt.Sprintf("/v1/users/%s", sd.Users[0].ID),
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
			URL:        fmt.Sprintf("/v1/users/%s", sd.Admins[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Model: &usergrp.AppUpdateUser{
				Name:            dbtest.StringPointer("Bill Kennedy"),
				Email:           dbtest.StringPointer("bill@ardanlabs.com"),
				Roles:           []string{"ADMIN"},
				Department:      dbtest.StringPointer("IT"),
				Password:        dbtest.StringPointer("123"),
				PasswordConfirm: dbtest.StringPointer("123"),
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
