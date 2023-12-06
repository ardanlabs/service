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

	sd, err := createUserSeed(context.Background(), dbTest)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	app.test(t, userQuery200(t, app, sd), "user-query-200")
	app.test(t, userQueryByID200(t, app, sd), "user-querybyid-200")

	app.test(t, userCreate200(t, app, sd), "user-create-200")
	app.test(t, userCreate401(t, app, sd), "user-create-2401")
	app.test(t, userCreate400(t, app, sd), "user-create-2400")

	app.test(t, userUpdate200(t, app, sd), "user-update-200")
	app.test(t, userUpdate401(t, app, sd), "user-update-401")
	app.test(t, userUpdate400(t, app, sd), "user-update-400")

	app.test(t, userDelete200(t, app, sd), "user-delete-200")
	app.test(t, userDelete401(t, app, sd), "user-delete-401")
}
