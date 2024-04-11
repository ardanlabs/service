package tests

import (
	"runtime/debug"
	"testing"
)

func Test_User(t *testing.T) {
	t.Parallel()

	// -------------------------------------------------------------------------

	dbTest, appTest := startTest(t, "Test_User")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		dbTest.Teardown()
	}()

	// -------------------------------------------------------------------------

	sd, err := insertUserSeed(dbTest)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	appTest.Run(t, userQuery200(sd), "user-query-200")
	appTest.Run(t, userQueryByID200(sd), "user-querybyid-200")

	appTest.Run(t, userCreate200(sd), "user-create-200")
	appTest.Run(t, userCreate401(sd), "user-create-401")
	appTest.Run(t, userCreate400(sd), "user-create-400")

	appTest.Run(t, userUpdate200(sd), "user-update-200")
	appTest.Run(t, userUpdate401(sd), "user-update-401")
	appTest.Run(t, userUpdate400(sd), "user-update-400")

	appTest.Run(t, userDelete200(sd), "user-delete-200")
	appTest.Run(t, userDelete401(sd), "user-delete-401")
}
