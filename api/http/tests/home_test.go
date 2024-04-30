package tests

import (
	"runtime/debug"
	"testing"
)

func Test_Home(t *testing.T) {
	t.Parallel()

	// -------------------------------------------------------------------------

	test := startTest(t, "Test_Home")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.DB.Teardown()
	}()

	// -------------------------------------------------------------------------

	sd, err := insertHomeSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	test.Run(t, homeQuery200(sd), "home-query-200")
	test.Run(t, homeQueryByID200(sd), "home-querybyid-200")

	test.Run(t, homeCreate200(sd), "home-create-200")
	test.Run(t, homeCreate401(sd), "home-create-401")
	test.Run(t, homeCreate400(sd), "home-create-400")

	test.Run(t, homeUpdate200(sd), "home-update-200")
	test.Run(t, homeUpdate401(sd), "home-update-401")
	test.Run(t, homeUpdate400(sd), "home-update-400")

	test.Run(t, homeDelete200(sd), "home-delete-200")
	test.Run(t, homeDelete401(sd), "home-delete-401")
}
