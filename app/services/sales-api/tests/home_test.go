package tests

import (
	"os"
	"runtime/debug"
	"testing"

	"github.com/ardanlabs/service/app/services/sales-api/build/all"
	"github.com/ardanlabs/service/business/api/mux"
	"github.com/ardanlabs/service/business/data/dbtest"
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

	app := dbtest.AppTest{
		Handler: mux.WebAPI(mux.Config{
			Shutdown: make(chan os.Signal, 1),
			Log:      dbTest.Log,
			Delegate: dbTest.Core.Delegate,
			Auth:     dbTest.Auth,
			DB:       dbTest.DB,
		}, all.Routes()),
	}

	// -------------------------------------------------------------------------

	sd, err := insertHomeSeed(dbTest)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	app.Test(t, homeQuery200(sd), "home-query-200")
	app.Test(t, homeQueryByID200(sd), "home-querybyid-200")

	app.Test(t, homeCreate200(sd), "home-create-200")
	app.Test(t, homeCreate401(sd), "home-create-401")
	app.Test(t, homeCreate400(sd), "home-create-400")

	app.Test(t, homeUpdate200(sd), "home-update-200")
	app.Test(t, homeUpdate401(sd), "home-update-401")
	app.Test(t, homeUpdate400(sd), "home-update-400")

	app.Test(t, homeDelete200(sd), "home-delete-200")
	app.Test(t, homeDelete401(sd), "home-delete-401")
}
