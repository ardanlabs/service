package tests

import (
	"context"
	"os"
	"runtime/debug"
	"testing"

	"github.com/ardanlabs/service/app/services/sales-api/v1/build/all"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/web/v1/mux"
)

func Test_Home(t *testing.T) {
	t.Parallel()

	dbTest := dbtest.NewTest(t, c, "Test_Home")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		dbTest.Teardown()
	}()

	app := appTest{
		Handler: mux.WebAPI(mux.Config{
			Shutdown: make(chan os.Signal, 1),
			Log:      dbTest.Log,
			Auth:     dbTest.V1.Auth,
			DB:       dbTest.DB,
		}, all.Routes()),
		userToken:  dbTest.TokenV1("user@example.com", "gophers"),
		adminToken: dbTest.TokenV1("admin@example.com", "gophers"),
	}

	// -------------------------------------------------------------------------

	sd, err := createHomeSeed(context.Background(), dbTest)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	app.test(t, homeQuery200(t, app, sd), "home-query-200")
	app.test(t, homeQueryByID200(t, app, sd), "home-querybyid-200")

	app.test(t, homeCreate200(t, app, sd), "home-create-200")
	app.test(t, homeCreate401(t, app, sd), "home-create-401")
	app.test(t, homeCreate400(t, app, sd), "home-create-400")

	app.test(t, homeUpdate200(t, app, sd), "home-update-200")
	app.test(t, homeUpdate401(t, app, sd), "home-update-401")
	app.test(t, homeUpdate400(t, app, sd), "home-update-400")

	app.test(t, homeDelete200(t, app, sd), "home-delete-200")
	app.test(t, homeDelete401(t, app, sd), "home-delete-401")
}
