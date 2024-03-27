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

func userUpdate200(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "basic",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodPut,
			statusCode: http.StatusOK,
			model: &usergrp.AppUpdateUser{
				Name:            dbtest.StringPointer("Jack Kennedy"),
				Email:           dbtest.StringPointer("jack@ardanlabs.com"),
				Roles:           []string{"ADMIN"},
				Department:      dbtest.StringPointer("IT"),
				Password:        dbtest.StringPointer("123"),
				PasswordConfirm: dbtest.StringPointer("123"),
			},
			resp: &usergrp.AppUser{},
			expResp: &usergrp.AppUser{
				ID:          sd.users[0].ID.String(),
				Name:        "Jack Kennedy",
				Email:       "jack@ardanlabs.com",
				Roles:       []string{"ADMIN"},
				Department:  "IT",
				Enabled:     true,
				DateCreated: sd.users[0].DateCreated.Format(time.RFC3339),
				DateUpdated: sd.users[0].DateUpdated.Format(time.RFC3339),
			},
			cmpFunc: func(got any, exp any) string {
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

func userUpdate400(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "bad-input",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodPut,
			statusCode: http.StatusBadRequest,
			model: &usergrp.AppUpdateUser{
				Email:           dbtest.StringPointer("bill@"),
				PasswordConfirm: dbtest.StringPointer("jack"),
			},
			resp: &errs.Response{},
			expResp: &errs.Response{
				Error:  "data validation error",
				Fields: map[string]string{"email": "email must be a valid email address", "passwordConfirm": "passwordConfirm must be equal to Password"},
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			name:       "bad-role",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodPut,
			statusCode: http.StatusBadRequest,
			model: &usergrp.AppUpdateUser{
				Roles: []string{"BAD ROLE"},
			},
			resp: &errs.Response{},
			expResp: &errs.Response{
				Error: "parse: invalid role \"BAD ROLE\"",
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func userUpdate401(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "emptytoken",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
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
			url:        fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
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
			url:        fmt.Sprintf("/v1/users/%s", sd.admins[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodPut,
			statusCode: http.StatusUnauthorized,
			model: &usergrp.AppUpdateUser{
				Name:            dbtest.StringPointer("Bill Kennedy"),
				Email:           dbtest.StringPointer("bill@ardanlabs.com"),
				Roles:           []string{"ADMIN"},
				Department:      dbtest.StringPointer("IT"),
				Password:        dbtest.StringPointer("123"),
				PasswordConfirm: dbtest.StringPointer("123"),
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
