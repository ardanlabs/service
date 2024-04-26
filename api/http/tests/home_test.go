package tests

import (
	"runtime/debug"
	"testing"
)

func Test_Home(t *testing.T) {
	t.Parallel()

	// -------------------------------------------------------------------------

	appTest := startTest(t, "Test_Home")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		appTest.dbTest.Teardown()
	}()

	// -------------------------------------------------------------------------

	sd, err := insertHomeSeed(appTest.dbTest, appTest.auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	appTest.run(t, homeQuery200(sd), "home-query-200")
	appTest.run(t, homeQueryByID200(sd), "home-querybyid-200")

	appTest.run(t, homeCreate200(sd), "home-create-200")
	appTest.run(t, homeCreate401(sd), "home-create-401")
	appTest.run(t, homeCreate400(sd), "home-create-400")

	appTest.run(t, homeUpdate200(sd), "home-update-200")
	appTest.run(t, homeUpdate401(sd), "home-update-401")
	appTest.run(t, homeUpdate400(sd), "home-update-400")

	appTest.run(t, homeDelete200(sd), "home-delete-200")
	appTest.run(t, homeDelete401(sd), "home-delete-401")
}
