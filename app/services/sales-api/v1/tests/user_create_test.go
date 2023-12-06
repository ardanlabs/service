package tests

import (
	"net/http"
	"testing"

	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/usergrp"
	v1 "github.com/ardanlabs/service/business/web/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func userCreate200(t *testing.T, app appTest, sd seedData) []tableData {
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
			cmpFunc: func(x interface{}, y interface{}) string {
				resp := x.(*usergrp.AppUser)
				expResp := y.(*usergrp.AppUser)

				if _, err := uuid.Parse(resp.ID); err != nil {
					return "bad uuid for ID"
				}

				if resp.DateCreated == "" {
					return "missing date created"
				}

				if resp.DateUpdated == "" {
					return "missing date updated"
				}

				expResp.ID = resp.ID
				expResp.DateCreated = resp.DateCreated
				expResp.DateUpdated = resp.DateUpdated

				return cmp.Diff(x, y)
			},
		},
	}

	return table
}

func userCreate400(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "missing-input",
			url:        "/v1/users",
			token:      sd.admins[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusBadRequest,
			model:      &usergrp.AppNewUser{},
			resp:       &v1.ErrorResponse{},
			expResp: &v1.ErrorResponse{
				Error:  "data validation error",
				Fields: map[string]string{"email": "email is a required field", "name": "name is a required field", "password": "password is a required field", "roles": "roles is a required field"},
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
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
			resp: &v1.ErrorResponse{},
			expResp: &v1.ErrorResponse{
				Error: "parse: invalid role \"BAD ROLE\"",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}

func userCreate401(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "emptytoken",
			url:        "/v1/users",
			token:      "",
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &v1.ErrorResponse{},
			expResp: &v1.ErrorResponse{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "badtoken",
			url:        "/v1/users",
			token:      sd.admins[0].token[:10],
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &v1.ErrorResponse{},
			expResp: &v1.ErrorResponse{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "badsig",
			url:        "/v1/users",
			token:      sd.admins[0].token + "A",
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &v1.ErrorResponse{},
			expResp: &v1.ErrorResponse{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "wronguser",
			url:        "/v1/users",
			token:      sd.users[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &v1.ErrorResponse{},
			expResp: &v1.ErrorResponse{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}
