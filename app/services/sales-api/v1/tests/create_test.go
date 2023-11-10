package tests

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"testing"

	"github.com/ardanlabs/service/app/services/sales-api/v1/build/all"
	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/homegrp"
	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/productgrp"
	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/usergrp"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/data/order"
	v1 "github.com/ardanlabs/service/business/web/v1"
	"github.com/ardanlabs/service/business/web/v1/response"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func createTests(t *testing.T, app appTest, sd seedData) {
	app.test(t, testCreate200(t, app, sd), "create200")
	app.test(t, testCreate401(t, app, sd), "create401")
}

func testCreate200(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "user",
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
		{
			name:       "product",
			url:        "/v1/products",
			token:      sd.users[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusCreated,
			model: &productgrp.AppNewProduct{
				Name:     "Guitar",
				Cost:     10.34,
				Quantity: 10,
			},
			resp: &productgrp.AppProduct{},
			expResp: &productgrp.AppProduct{
				Name:     "Guitar",
				UserID:   sd.users[0].ID.String(),
				Cost:     10.34,
				Quantity: 10,
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				resp := x.(*productgrp.AppProduct)
				expResp := y.(*productgrp.AppProduct)

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
		{
			name:       "home",
			url:        "/v1/homes",
			token:      sd.users[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusCreated,
			model: &homegrp.AppNewHome{
				Type: "SINGLE FAMILY",
				Address: homegrp.AppNewAddress{
					Address1: "123 Mocking Bird Lane",
					ZipCode:  "35810",
					City:     "Huntsville",
					State:    "AL",
					Country:  "US",
				},
			},
			resp: &homegrp.AppHome{},
			expResp: &homegrp.AppHome{
				UserID: sd.users[0].ID.String(),
				Type:   "SINGLE FAMILY",
				Address: homegrp.AppNewAddress{
					Address1: "123 Mocking Bird Lane",
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

func testCreate401(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "user-emptytoken",
			url:        "/v1/users",
			token:      "",
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "user-badtoken",
			url:        "/v1/users",
			token:      sd.admins[0].token[:10],
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "user-badsig",
			url:        "/v1/users",
			token:      sd.admins[0].token + "A",
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "user-wronguser",
			url:        "/v1/users",
			token:      sd.users[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "product-emptytoken",
			url:        "/v1/products",
			token:      "",
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "product-badtoken",
			url:        "/v1/products",
			token:      sd.admins[0].token[:10],
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "product-badsig",
			url:        "/v1/products",
			token:      sd.admins[0].token + "A",
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "product-wronguser",
			url:        "/v1/products",
			token:      sd.admins[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "home-emptytoken",
			url:        "/v1/homes",
			token:      "",
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "home-badtoken",
			url:        "/v1/homes",
			token:      sd.admins[0].token[:10],
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "home-badsig",
			url:        "/v1/homes",
			token:      sd.admins[0].token + "A",
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "home-wronguser",
			url:        "/v1/homes",
			token:      sd.admins[0].token,
			method:     http.MethodPost,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp: &response.ErrorDocument{
				Error: "Unauthorized",
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}

// =============================================================================

func createSeed(ctx context.Context, dbTest *dbtest.Test) (seedData, error) {
	usrs, err := dbTest.CoreAPIs.User.Query(ctx, user.QueryFilter{}, order.By{Field: user.OrderByName, Direction: order.ASC}, 1, 2)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	// -------------------------------------------------------------------------

	tu1 := testUser{
		User:  usrs[0],
		token: dbTest.TokenV1(usrs[0].Email.Address, "gophers"),
	}

	tu2 := testUser{
		User:  usrs[1],
		token: dbTest.TokenV1(usrs[1].Email.Address, "gophers"),
	}

	// -------------------------------------------------------------------------

	sd := seedData{
		admins: []testUser{tu1},
		users:  []testUser{tu2},
	}

	return sd, nil
}

// =============================================================================

func Test_Create(t *testing.T) {
	t.Parallel()

	dbTest := dbtest.NewTest(t, c)
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		dbTest.Teardown()
	}()

	app := appTest{
		Handler: v1.APIMux(v1.APIMuxConfig{
			Shutdown: make(chan os.Signal, 1),
			Log:      dbTest.Log,
			Auth:     dbTest.V1.Auth,
			DB:       dbTest.DB,
		}, all.Routes()),
		userToken:  dbTest.TokenV1("user@example.com", "gophers"),
		adminToken: dbTest.TokenV1("admin@example.com", "gophers"),
	}

	// -------------------------------------------------------------------------

	t.Log("Seeding data ...")
	sd, err := createSeed(context.Background(), dbTest)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	createTests(t, app, sd)
}
