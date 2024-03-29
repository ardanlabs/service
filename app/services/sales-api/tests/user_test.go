package tests

import (
	"os"
	"runtime/debug"
	"testing"

	"github.com/ardanlabs/service/app/services/sales-api/build/all"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/web/mux"
)

func Test_User(t *testing.T) {
	t.Parallel()

	dbTest := dbtest.NewTest(t, c, "Test_User")
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

	sd, err := insertUserSeed(dbTest)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	app.Test(t, userQuery200(sd), "user-query-200")
	app.Test(t, userQueryByID200(sd), "user-querybyid-200")

	app.Test(t, userCreate200(sd), "user-create-200")
	app.Test(t, userCreate401(sd), "user-create-401")
	app.Test(t, userCreate400(sd), "user-create-400")

	app.Test(t, userUpdate200(sd), "user-update-200")
	app.Test(t, userUpdate401(sd), "user-update-401")
	app.Test(t, userUpdate400(sd), "user-update-400")

	app.Test(t, userDelete200(sd), "user-delete-200")
	app.Test(t, userDelete401(sd), "user-delete-401")
}
