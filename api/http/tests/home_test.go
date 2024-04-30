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
		apiTest.DB.Teardown()
	}()

	// -------------------------------------------------------------------------

	sd, err := insertHomeSeed(apiTest.DB, apiTest.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	apiTest.Run(t, homeQuery200(sd), "home-query-200")
	apiTest.Run(t, homeQueryByID200(sd), "home-querybyid-200")

	apiTest.Run(t, homeCreate200(sd), "home-create-200")
	apiTest.Run(t, homeCreate401(sd), "home-create-401")
	apiTest.Run(t, homeCreate400(sd), "home-create-400")

	apiTest.Run(t, homeUpdate200(sd), "home-update-200")
	apiTest.Run(t, homeUpdate401(sd), "home-update-401")
	apiTest.Run(t, homeUpdate400(sd), "home-update-400")

	apiTest.Run(t, homeDelete200(sd), "home-delete-200")
	apiTest.Run(t, homeDelete401(sd), "home-delete-401")
}
