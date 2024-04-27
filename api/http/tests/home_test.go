package tests

import (
	"runtime/debug"
	"testing"
)

func Test_Home(t *testing.T) {
	t.Parallel()

	// -------------------------------------------------------------------------

	apiTest := startTest(t, "Test_Home")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		apiTest.dbTest.Teardown()
	}()

	// -------------------------------------------------------------------------

	sd, err := insertHomeSeed(apiTest.dbTest, apiTest.auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	apiTest.run(t, homeQuery200(sd), "home-query-200")
	apiTest.run(t, homeQueryByID200(sd), "home-querybyid-200")

	apiTest.run(t, homeCreate200(sd), "home-create-200")
	apiTest.run(t, homeCreate401(sd), "home-create-401")
	apiTest.run(t, homeCreate400(sd), "home-create-400")

	apiTest.run(t, homeUpdate200(sd), "home-update-200")
	apiTest.run(t, homeUpdate401(sd), "home-update-401")
	apiTest.run(t, homeUpdate400(sd), "home-update-400")

	apiTest.run(t, homeDelete200(sd), "home-delete-200")
	apiTest.run(t, homeDelete401(sd), "home-delete-401")
}
