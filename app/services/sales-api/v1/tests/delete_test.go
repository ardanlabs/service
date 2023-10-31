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
)

func deleteTests(t *testing.T, tests appTest, sd seedData) {
	tests.run(t, testDelete200(t, sd), "delete200")
}

func testDelete200(t *testing.T, sd seedData) []tableData {
	table := []tableData{
		{
			name: "product",
			url:  fmt.Sprintf("/v1/products/%s", sd.products[0].ID),
		},
		{
			name: "home",
			url:  fmt.Sprintf("/v1/homes/%s", sd.homes[0].ID),
		},
		{
			name: "user",
			url:  fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
		},
	}

	return table
}

// =============================================================================

func deleteSeed(ctx context.Context, api dbtest.CoreAPIs) (seedData, error) {
	usrs, err := user.TestGenerateSeedUsers(1, api.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding products : %w", err)
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

func Test_Delete(t *testing.T) {
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
		method:     http.MethodDelete,
		statusCode: http.StatusNoContent,
		userToken:  test.TokenV1("user@example.com", "gophers"),
		adminToken: test.TokenV1("admin@example.com", "gophers"),
	}

	// -------------------------------------------------------------------------

	t.Log("Seeding data ...")
	sd, err := deleteSeed(context.Background(), test.CoreAPIs)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	deleteTests(t, tests, sd)
}
