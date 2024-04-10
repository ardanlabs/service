package tests

import (
	"os"
	"runtime/debug"
	"testing"

	"github.com/ardanlabs/service/apis/services/sales/http/build/all"
	"github.com/ardanlabs/service/app/api/apptest"
	"github.com/ardanlabs/service/app/api/mux"
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

	app := apptest.New(mux.WebAPI(mux.Config{
		Shutdown: make(chan os.Signal, 1),
		Log:      dbTest.Log,
		Auth:     dbTest.Auth,
		DB:       dbTest.DB,
		BusCrud: mux.BusCrud{
			Delegate: dbTest.Core.BusCrud.Delegate,
			Home:     dbTest.Core.BusCrud.Home,
			Product:  dbTest.Core.BusCrud.Product,
			User:     dbTest.Core.BusCrud.User,
		},
		BusView: mux.BusView{
			Product: dbTest.Core.BusView.Product,
		},
	}, all.Routes()))

	// -------------------------------------------------------------------------

	sd, err := insertVProductSeed(dbTest)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	app.Test(t, vproductQuery200(sd), "vproduct-query-200")
}
