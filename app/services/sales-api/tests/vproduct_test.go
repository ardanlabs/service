package tests

import (
	"os"
	"runtime/debug"
	"testing"

	"github.com/ardanlabs/service/app/services/sales-api/build/all"
	"github.com/ardanlabs/service/business/api/mux"
	"github.com/ardanlabs/service/business/data/dbtest"
)

func Test_VProduct(t *testing.T) {
	t.Parallel()

	dbTest := dbtest.NewTest(t, c, "Test_VProduct")
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
			Auth:     dbTest.Auth,
			DB:       dbTest.DB,
		}, all.Routes()),
	}

	// -------------------------------------------------------------------------

	sd, err := insertVProductSeed(dbTest)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	app.Test(t, vproductQuery200(sd), "vproduct-query-200")
}
