package tests

import (
	"os"
	"runtime/debug"
	"testing"

	"github.com/ardanlabs/service/app/services/sales-api/build/all"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/web/mux"
)

func Test_Product(t *testing.T) {
	t.Parallel()

	dbTest := dbtest.NewTest(t, c, "Test_Product")
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

	sd, err := insertProductSeed(dbTest)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	app.Test(t, productQuery200(sd), "product-query-200")
	app.Test(t, productQueryByID200(sd), "product-querybyid-200")

	app.Test(t, productCreate200(sd), "product-create-200")
	app.Test(t, productCreate401(sd), "product-create-401")
	app.Test(t, productCreate400(sd), "product-create-400")

	app.Test(t, productUpdate200(sd), "product-update-200")
	app.Test(t, productUpdate401(sd), "product-update-401")
	app.Test(t, productUpdate400(sd), "product-update-400")

	app.Test(t, productDelete200(sd), "product-delete-200")
	app.Test(t, productDelete401(sd), "product-delete-401")
}
