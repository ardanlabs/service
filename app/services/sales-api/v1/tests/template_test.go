package tests

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"testing"

	"github.com/ardanlabs/service/app/services/sales-api/v1/build/all"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/data/order"
	v1 "github.com/ardanlabs/service/business/web/v1"
)

func nameTests(t *testing.T, app appTest, sd seedData) {
	app.test(t, testName200(t, app, sd), "name200")
}

func testName200(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{}

	return table
}

// =============================================================================

func nameSeed(ctx context.Context, dbTest *dbtest.Test) (seedData, error) {
	usrs, err := dbTest.CoreAPIs.User.Query(ctx, user.QueryFilter{}, order.By{Field: user.OrderByName, Direction: order.ASC}, 1, 2)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tkns := make([]string, len(usrs))
	for _, u := range usrs {
		tkns = append(tkns, dbTest.TokenV1(u.Email.Address, "gophers"))
	}

	sd := seedData{
		tokens: tkns,
		users:  usrs,
	}

	return sd, nil
}

// =============================================================================

func Test_Name(t *testing.T) {
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
	}

	// -------------------------------------------------------------------------

	t.Log("Seeding data ...")
	sd, err := nameSeed(context.Background(), dbTest)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	nameTests(t, app, sd)
}
