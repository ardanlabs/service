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

func createTests(t *testing.T, app appTest, sd seedData) {
	app.test(t, testCreate200(t, app, sd), "create200")
	app.test(t, testCreate401(t, app, sd), "create401")
}

// =============================================================================

func Test_Create(t *testing.T) {
	t.Parallel()

	dbTest := dbtest.NewTest(t, c)
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

	t.Log("Seeding data ...")
	sd, err := createSeed(context.Background(), dbTest)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	createTests(t, app, sd)
}
