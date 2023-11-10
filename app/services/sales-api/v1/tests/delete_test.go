package tests

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"testing"

	"github.com/ardanlabs/service/app/services/sales-api/v1/build/all"
	"github.com/ardanlabs/service/business/core/home"
	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	v1 "github.com/ardanlabs/service/business/web/v1"
	"github.com/ardanlabs/service/business/web/v1/response"
	"github.com/google/go-cmp/cmp"
)

func deleteTests(t *testing.T, app appTest, sd seedData) {
	app.test(t, testDelete200(t, app, sd), "delete200")
	app.test(t, testDelete401(t, app, sd), "delete401")
}

func testDelete200(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "product-user",
			url:        fmt.Sprintf("/v1/products/%s", sd.users[0].products[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "product-admin",
			url:        fmt.Sprintf("/v1/products/%s", sd.admins[0].products[0].ID),
			token:      sd.admins[0].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "home-user",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "home-admin",
			url:        fmt.Sprintf("/v1/homes/%s", sd.admins[0].homes[0].ID),
			token:      sd.admins[0].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "user-user",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[1].ID),
			token:      sd.users[1].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "user-admin",
			url:        fmt.Sprintf("/v1/users/%s", sd.admins[1].ID),
			token:      sd.admins[1].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
	}

	return table
}

func testDelete401(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "product-emptytoken",
			url:        fmt.Sprintf("/v1/products/%s", sd.users[0].products[1].ID),
			token:      "",
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp:    &response.ErrorDocument{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "product-badsig",
			url:        fmt.Sprintf("/v1/products/%s", sd.users[0].products[1].ID),
			token:      sd.users[0].token + "A",
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp:    &response.ErrorDocument{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "product-wronguser",
			url:        fmt.Sprintf("/v1/products/%s", sd.users[0].products[1].ID),
			token:      app.userToken,
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp:    &response.ErrorDocument{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "home-emptytoken",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[1].ID),
			token:      "",
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp:    &response.ErrorDocument{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "home-badsig",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[1].ID),
			token:      sd.users[0].token + "A",
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp:    &response.ErrorDocument{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "home-wronguser",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[1].ID),
			token:      app.userToken,
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp:    &response.ErrorDocument{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "user-emptytoken",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
			token:      "",
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp:    &response.ErrorDocument{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "user-badsig",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
			token:      sd.users[0].token + "A",
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp:    &response.ErrorDocument{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "user-wronguser",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
			token:      app.userToken,
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp:    &response.ErrorDocument{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}

// =============================================================================

func deleteSeed(ctx context.Context, dbTest *dbtest.Test) (seedData, error) {
	usrs, err := user.TestGenerateSeedUsers(1, user.RoleUser, dbTest.CoreAPIs.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	prds, err := product.TestGenerateSeedProducts(2, dbTest.CoreAPIs.Product, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding products : %w", err)
	}

	hmes, err := home.TestGenerateSeedHomes(2, dbTest.CoreAPIs.Home, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding homes : %w", err)
	}

	tu1 := testUser{
		User:     usrs[0],
		token:    dbTest.TokenV1(usrs[0].Email.Address, fmt.Sprintf("Password%s", usrs[0].Name[4:])),
		products: prds,
		homes:    hmes,
	}

	// -------------------------------------------------------------------------

	usrs, err = user.TestGenerateSeedUsers(1, user.RoleUser, dbTest.CoreAPIs.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu2 := testUser{
		User:  usrs[0],
		token: dbTest.TokenV1(usrs[0].Email.Address, fmt.Sprintf("Password%s", usrs[0].Name[4:])),
	}

	// -------------------------------------------------------------------------

	usrs, err = user.TestGenerateSeedUsers(1, user.RoleAdmin, dbTest.CoreAPIs.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	prds, err = product.TestGenerateSeedProducts(2, dbTest.CoreAPIs.Product, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding products : %w", err)
	}

	hmes, err = home.TestGenerateSeedHomes(2, dbTest.CoreAPIs.Home, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding homes : %w", err)
	}

	tu3 := testUser{
		User:     usrs[0],
		token:    dbTest.TokenV1(usrs[0].Email.Address, fmt.Sprintf("Password%s", usrs[0].Name[4:])),
		products: prds,
		homes:    hmes,
	}

	// -------------------------------------------------------------------------

	usrs, err = user.TestGenerateSeedUsers(1, user.RoleAdmin, dbTest.CoreAPIs.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu4 := testUser{
		User:  usrs[0],
		token: dbTest.TokenV1(usrs[0].Email.Address, fmt.Sprintf("Password%s", usrs[0].Name[4:])),
	}

	// -------------------------------------------------------------------------

	sd := seedData{
		users:  []testUser{tu1, tu2},
		admins: []testUser{tu3, tu4},
	}

	return sd, nil
}

// =============================================================================

func Test_Delete(t *testing.T) {
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
	sd, err := deleteSeed(context.Background(), dbTest)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	deleteTests(t, app, sd)
}
