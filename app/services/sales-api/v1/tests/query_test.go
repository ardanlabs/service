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
	"github.com/ardanlabs/service/business/web/v1/response"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func queryTests(t *testing.T, app appTest, sd seedData) {
	app.test(t, testQuery200(t, app, sd), "query200")
	app.test(t, testQueryByID200(t, app, sd), "queryByID200")
}

func testQuery200(t *testing.T, app appTest, sd seedData) []tableData {
	usrs := make([]user.User, 0, len(sd.admins)+len(sd.users))
	usrsMap := make(map[uuid.UUID]user.User)
	for _, adm := range sd.admins {
		usrsMap[adm.ID] = adm.User
		usrs = append(usrs, adm.User)
	}
	for _, usr := range sd.users {
		usrsMap[usr.ID] = usr.User
		usrs = append(usrs, usr.User)
	}

	table := []tableData{
		{
			name:       "user",
			url:        "/v1/users?page=1&rows=2&orderBy=user_id,DESC",
			token:      sd.admins[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &response.PageDocument[usergrp.AppUser]{},
			expResp: &response.PageDocument[usergrp.AppUser]{
				Page:        1,
				RowsPerPage: 2,
				Total:       len(usrs),
				Items:       toAppUsers(usrs),
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "product",
			url:        "/v1/products?page=1&rows=10&orderBy=user_id,DESC",
			token:      sd.admins[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &response.PageDocument[productgrp.AppProductDetails]{},
			expResp: &response.PageDocument[productgrp.AppProductDetails]{
				Page:        1,
				RowsPerPage: 10,
				Total:       len(sd.admins[0].products) + len(sd.users[0].products),
				Items:       toAppProductsDetails(append(sd.admins[0].products, sd.users[0].products...), usrsMap),
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "home",
			url:        "/v1/homes?page=1&rows=10&orderBy=user_id,DESC",
			token:      sd.admins[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &response.PageDocument[homegrp.AppHome]{},
			expResp: &response.PageDocument[homegrp.AppHome]{
				Page:        1,
				RowsPerPage: 10,
				Total:       len(sd.admins[0].homes) + len(sd.users[0].homes),
				Items:       toAppHomes(append(sd.admins[0].homes, sd.users[0].homes...)),
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}

func testQueryByID200(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "user",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
			token:      sd.users[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &usergrp.AppUser{},
			expResp:    toAppUserPtr(sd.users[0].User),
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "product",
			url:        fmt.Sprintf("/v1/products/%s", sd.users[0].products[0].ID),
			token:      sd.users[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &productgrp.AppProduct{},
			expResp:    toAppProductPtr(sd.users[0].products[0]),
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "home",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[0].ID),
			token:      sd.users[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &homegrp.AppHome{},
			expResp:    toAppHomePtr(sd.users[0].homes[0]),
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}

// =============================================================================

func querySeed(ctx context.Context, dbTest *dbtest.Test) (seedData, error) {
	usrs, err := dbTest.CoreAPIs.User.Query(ctx, user.QueryFilter{}, order.By{Field: user.OrderByName, Direction: order.ASC}, 1, 2)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	// -------------------------------------------------------------------------

	tu1 := testUser{
		User:  usrs[0],
		token: dbTest.TokenV1(usrs[0].Email.Address, "gophers"),
	}

	tu1.products, err = product.TestGenerateSeedProducts(5, dbTest.CoreAPIs.Product, tu1.ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding products1 : %w", err)
	}

	tu1.homes, err = home.TestGenerateSeedHomes(5, dbTest.CoreAPIs.Home, tu1.ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding homes1 : %w", err)
	}

	// -------------------------------------------------------------------------

	tu2 := testUser{
		User:  usrs[1],
		token: dbTest.TokenV1(usrs[1].Email.Address, "gophers"),
	}

	tu2.products, err = product.TestGenerateSeedProducts(5, dbTest.CoreAPIs.Product, tu2.ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding products2 : %w", err)
	}

	tu2.homes, err = home.TestGenerateSeedHomes(5, dbTest.CoreAPIs.Home, tu2.ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding homes2 : %w", err)
	}

	// -------------------------------------------------------------------------

	sd := seedData{
		admins: []testUser{tu1},
		users:  []testUser{tu2},
	}

	return sd, nil
}

// =============================================================================

func Test_Query(t *testing.T) {
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
	sd, err := querySeed(context.Background(), dbTest)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	queryTests(t, app, sd)
}
