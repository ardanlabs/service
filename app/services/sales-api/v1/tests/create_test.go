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

func Test_Create(t *testing.T) {
	t.Parallel()

	dbTest := dbtest.NewTest(t, c, "Test_Create")
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

	sd, err := createSeed(context.Background(), dbTest)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	app.test(t, userCreate200(t, app, sd), "user-200")
	app.test(t, productCreate200(t, app, sd), "product-200")
	app.test(t, homeCreate200(t, app, sd), "home-200")

	app.test(t, userCreate401(t, app, sd), "user-401")
	app.test(t, productCreate401(t, app, sd), "product-401")
	app.test(t, homeCreate401(t, app, sd), "home-401")

	app.test(t, userCreate400(t, app, sd), "user-400")
	app.test(t, productCreate400(t, app, sd), "product-400")
	app.test(t, homeCreate400(t, app, sd), "home-400")
}
