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
	"github.com/ardanlabs/service/business/core/home"
	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/data/order"
	v1 "github.com/ardanlabs/service/business/web/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func updateTests(t *testing.T, tests appTest, sd seedData) {
	tests.run(t, testUpdate200(t, sd), "update200")
}

func testUpdate200(t *testing.T, sd seedData) []tableData {
	table := []tableData{
		{
			name: "user",
			url:  fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
			model: &usergrp.AppUpdateUser{
				Name:            dbtest.StringPointer("Bill Kennedy"),
				Email:           dbtest.StringPointer("bill@ardanlabs.com"),
				Roles:           []string{"ADMIN"},
				Department:      dbtest.StringPointer("IT"),
				Password:        dbtest.StringPointer("123"),
				PasswordConfirm: dbtest.StringPointer("123"),
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
			name: "product",
			url:  fmt.Sprintf("/v1/products/%s", sd.products[0].ID),
			model: &productgrp.AppUpdateProduct{
				Name:     dbtest.StringPointer("Guitar"),
				Cost:     dbtest.FloatPointer(10.34),
				Quantity: dbtest.IntPointer(10),
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
			name: "home",
			url:  fmt.Sprintf("/v1/homes/%s", sd.homes[0].ID),
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
				Address: homegrp.AppNewAddress{
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

// =============================================================================

func updateSeed(ctx context.Context, api dbtest.CoreAPIs) (seedData, error) {
	usrs, err := api.User.Query(ctx, user.QueryFilter{}, order.By{Field: user.OrderByName, Direction: order.ASC}, 1, 2)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	prds, err := product.TestGenerateSeedProducts(1, api.Product, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding products : %w", err)
	}

	hmes, err := home.TestGenerateSeedHomes(1, api.Home, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding homes : %w", err)
	}

	sd := seedData{
		users:    usrs,
		products: prds,
		homes:    hmes,
	}

	return sd, nil
}

// =============================================================================

func Test_Update(t *testing.T) {
	t.Parallel()

	test := dbtest.NewTest(t, c)
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.Teardown()
	}()

	tests := appTest{
		app: v1.APIMux(v1.APIMuxConfig{
			Shutdown: make(chan os.Signal, 1),
			Log:      test.Log,
			Auth:     test.V1.Auth,
			DB:       test.DB,
		}, all.Routes()),
		method:     http.MethodPut,
		statusCode: http.StatusOK,
		userToken:  test.TokenV1("user@example.com", "gophers"),
		adminToken: test.TokenV1("admin@example.com", "gophers"),
	}

	// -------------------------------------------------------------------------

	t.Log("Seeding data ...")
	sd, err := updateSeed(context.Background(), test.CoreAPIs)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	updateTests(t, tests, sd)
}
