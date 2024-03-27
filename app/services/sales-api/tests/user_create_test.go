package tests

import (
	"net/http"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/crud/usergrp"
	"github.com/ardanlabs/service/business/web/errs"
	"github.com/google/go-cmp/cmp"
)

func userCreate200(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "basic",
			url:        "/v1/users",
			token:      sd.admins[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusCreated,
			model: &usergrp.AppNewUser{
				Name:            "Bill Kennedy",
				Email:           "bill@ardanlabs.com",
				Roles:           []string{"ADMIN"},
				Department:      "IT",
				Password:        "123",
				PasswordConfirm: "123",
			},
			resp: &usergrp.AppUser{},
			expResp: &usergrp.AppUser{
				Name:       "Bill Kennedy",
				Email:      "bill@ardanlabs.com",
				Roles:      []string{"ADMIN"},
				Department: "IT",
				Enabled:    true,
			},
			cmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*usergrp.AppUser)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*usergrp.AppUser)

				expResp.ID = gotResp.ID
				expResp.DateCreated = gotResp.DateCreated
				expResp.DateUpdated = gotResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func userCreate400(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "missing-input",
			url:        "/v1/users",
			token:      sd.admins[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusBadRequest,
			model:      &usergrp.AppNewUser{},
			resp:       &errs.Response{},
			expResp: &errs.Response{
				Error:  "data validation error",
				Fields: map[string]string{"email": "email is a required field", "name": "name is a required field", "password": "password is a required field", "roles": "roles is a required field"},
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			name:       "bad-role",
			url:        "/v1/users",
			token:      sd.admins[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusBadRequest,
			model: &usergrp.AppNewUser{
				Name:            "Bill Kennedy",
				Email:           "bill@ardanlabs.com",
				Roles:           []string{"BAD ROLE"},
				Department:      "IT",
				Password:        "123",
				PasswordConfirm: "123",
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

func userCreate401(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "emptytoken",
			url:        "/v1/users",
			token:      "",
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &errs.Response{},
			expResp: &errs.Response{
				Error: "Unauthorized",
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			name:       "badtoken",
			url:        "/v1/users",
			token:      sd.admins[0].token[:10],
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &errs.Response{},
			expResp: &errs.Response{
				Error: "Unauthorized",
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			name:       "badsig",
			url:        "/v1/users",
			token:      sd.admins[0].token + "A",
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &errs.Response{},
			expResp: &errs.Response{
				Error: "Unauthorized",
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			name:       "wronguser",
			url:        "/v1/users",
			token:      sd.users[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &errs.Response{},
			expResp: &errs.Response{
				Error: "Unauthorized",
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
