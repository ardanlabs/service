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
			url:        fmt.Sprintf("/v1/products/%s", sd.products[0].ID),
			token:      sd.tokens[0],
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "product-admin",
			url:        fmt.Sprintf("/v1/products/%s", sd.products[1].ID),
			token:      sd.tokens[1],
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "home-user",
			url:        fmt.Sprintf("/v1/homes/%s", sd.homes[0].ID),
			token:      sd.tokens[0],
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "home-admin",
			url:        fmt.Sprintf("/v1/homes/%s", sd.homes[1].ID),
			token:      sd.tokens[1],
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "user-user",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
			token:      sd.tokens[0],
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "user-admin",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[1].ID),
			token:      sd.tokens[1],
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
	}

	return table
}

func testDelete401(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "product",
			url:        fmt.Sprintf("/v1/products/%s", sd.products[0].ID),
			token:      sd.tokens[0] + "A",
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp:    &response.ErrorDocument{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "product",
			url:        fmt.Sprintf("/v1/products/%s", sd.products[0].ID),
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
			name:       "home",
			url:        fmt.Sprintf("/v1/homes/%s", sd.homes[0].ID),
			token:      sd.tokens[0] + "A",
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp:    &response.ErrorDocument{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "home",
			url:        fmt.Sprintf("/v1/homes/%s", sd.homes[0].ID),
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
			name:       "user",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
			token:      sd.tokens[0] + "A",
			method:     http.MethodDelete,
			statusCode: http.StatusUnauthorized,
			resp:       &response.ErrorDocument{},
			expResp:    &response.ErrorDocument{Error: "Unauthorized"},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "user",
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
	}

	return table
}

// =============================================================================

func deleteSeed(ctx context.Context, dbTest *dbtest.Test) (seedData, error) {
	usrs, err := user.TestGenerateSeedUsers(1, user.RoleUser, dbTest.CoreAPIs.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding products : %w", err)
	}

	usrs2, err := user.TestGenerateSeedUsers(1, user.RoleAdmin, dbTest.CoreAPIs.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding products : %w", err)
	}

	usrs = append(usrs, usrs2...)

	tkns := make([]string, 2)
	for i, usr := range usrs {
		tkns[i] = dbTest.TokenV1(usr.Email.Address, fmt.Sprintf("Password%s", usr.Name[4:]))
	}

	prds, err := product.TestGenerateSeedProducts(1, dbTest.CoreAPIs.Product, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding products : %w", err)
	}

	prds2, err := product.TestGenerateSeedProducts(1, dbTest.CoreAPIs.Product, usrs[1].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding products : %w", err)
	}

	prds = append(prds, prds2...)

	hmes, err := home.TestGenerateSeedHomes(1, dbTest.CoreAPIs.Home, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding homes : %w", err)
	}

	hmes2, err := home.TestGenerateSeedHomes(1, dbTest.CoreAPIs.Home, usrs[1].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding homes : %w", err)
	}

	hmes = append(hmes, hmes2...)

	sd := seedData{
		tokens:   tkns,
		users:    usrs,
		products: prds,
		homes:    hmes,
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
