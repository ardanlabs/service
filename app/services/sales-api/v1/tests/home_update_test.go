package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/homegrp"
	"github.com/ardanlabs/service/business/data/dbtest"
	v1 "github.com/ardanlabs/service/business/web/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func homeUpdate200(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "basic",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodPut,
			statusCode: http.StatusOK,
			model: &homegrp.AppUpdateHome{
				Type: dbtest.StringPointer("SINGLE FAMILY"),
				Address: &homegrp.AppUpdateAddress{
					Address1: dbtest.StringPointer("123 Mocking Bird Lane"),
					Address2: dbtest.StringPointer("apt 105"),
					ZipCode:  dbtest.StringPointer("35810"),
					City:     dbtest.StringPointer("Huntsville"),
					State:    dbtest.StringPointer("AL"),
					Country:  dbtest.StringPointer("US"),
				},
			},
			resp: &homegrp.AppHome{},
			expResp: &homegrp.AppHome{
				UserID: sd.users[0].ID.String(),
				Type:   "SINGLE FAMILY",
				Address: homegrp.AppAddress{
					Address1: "123 Mocking Bird Lane",
					Address2: "apt 105",
					ZipCode:  "35810",
					City:     "Huntsville",
					State:    "AL",
					Country:  "US",
				},
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				resp := x.(*homegrp.AppHome)
				expResp := y.(*homegrp.AppHome)

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

func homeUpdate400(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "bad-input",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodPut,
			statusCode: http.StatusBadRequest,
			model: &homegrp.AppUpdateHome{
				Address: &homegrp.AppUpdateAddress{
					Address1: dbtest.StringPointer(""),
					Address2: dbtest.StringPointer(""),
					ZipCode:  dbtest.StringPointer(""),
					City:     dbtest.StringPointer(""),
					State:    dbtest.StringPointer(""),
					Country:  dbtest.StringPointer(""),
				},
			},
			resp: &v1.ErrorResponse{},
			expResp: &v1.ErrorResponse{
				Error:  "data validation error",
				Fields: map[string]string{"address1": "address1 must be at least 1 character in length", "country": "Key: 'AppUpdateHome.address.country' Error:Field validation for 'country' failed on the 'iso3166_1_alpha2' tag", "state": "state must be at least 1 character in length", "zipCode": "zipCode must be a valid numeric value"},
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "bad-type",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodPut,
			statusCode: http.StatusBadRequest,
			model: &homegrp.AppUpdateHome{
				Type:    dbtest.StringPointer("BAD TYPE"),
				Address: &homegrp.AppUpdateAddress{},
			},
			resp: &v1.ErrorResponse{},
			expResp: &v1.ErrorResponse{
				Error: "parse: invalid type \"BAD TYPE\"",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}

func homeUpdate401(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "emptytoken",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[0].ID),
			token:      "",
			method:     http.MethodPut,
			statusCode: http.StatusUnauthorized,
			resp:       &v1.ErrorResponse{},
			expResp:    &v1.ErrorResponse{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "badsig",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[0].ID),
			token:      sd.users[0].token + "A",
			method:     http.MethodPut,
			statusCode: http.StatusUnauthorized,
			resp:       &v1.ErrorResponse{},
			expResp:    &v1.ErrorResponse{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "wronguser",
			url:        fmt.Sprintf("/v1/homes/%s", sd.admins[0].homes[0].ID),
			token:      app.userToken,
			method:     http.MethodPut,
			statusCode: http.StatusUnauthorized,
			model: &homegrp.AppUpdateHome{
				Type: dbtest.StringPointer("SINGLE FAMILY"),
				Address: &homegrp.AppUpdateAddress{
					Address1: dbtest.StringPointer("123 Mocking Bird Lane"),
					Address2: dbtest.StringPointer("apt 105"),
					ZipCode:  dbtest.StringPointer("35810"),
					City:     dbtest.StringPointer("Huntsville"),
					State:    dbtest.StringPointer("AL"),
					Country:  dbtest.StringPointer("US"),
				},
			},
			resp:    &v1.ErrorResponse{},
			expResp: &v1.ErrorResponse{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}
