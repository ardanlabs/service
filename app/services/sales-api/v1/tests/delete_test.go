package tests

import (
	"context"
	"os"
	"runtime/debug"
	"testing"

	"github.com/ardanlabs/service/app/services/sales-api/v1/build/all"
	"github.com/ardanlabs/service/business/data/dbtest"
	v1 "github.com/ardanlabs/service/business/web/v1"
)

func Test_Delete(t *testing.T) {
	t.Parallel()

	dbTest := dbtest.NewTest(t, c, "Test_Delete")
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

	sd, err := deleteSeed(context.Background(), dbTest)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	app.test(t, userDelete200(t, app, sd), "user-200")
	app.test(t, productDelete200(t, app, sd), "product-200")
	app.test(t, homeDelete200(t, app, sd), "home-200")

	app.test(t, userDelete401(t, app, sd), "user-401")
	app.test(t, productDelete401(t, app, sd), "product-401")
	app.test(t, homeDelete401(t, app, sd), "home-401")
}
