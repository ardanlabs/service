package tests

import (
	"runtime/debug"
	"testing"
)

func Test_User(t *testing.T) {
	t.Parallel()

	// -------------------------------------------------------------------------

	appTest := startTest(t, "Test_User")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		appTest.dbTest.Teardown()
	}()

	// -------------------------------------------------------------------------

	sd, err := insertUserSeed(appTest.dbTest, appTest.auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	appTest.run(t, userQuery200(sd), "user-query-200")
	appTest.run(t, userQueryByID200(sd), "user-querybyid-200")

	appTest.run(t, userCreate200(sd), "user-create-200")
	appTest.run(t, userCreate401(sd), "user-create-401")
	appTest.run(t, userCreate400(sd), "user-create-400")

	appTest.run(t, userUpdate200(sd), "user-update-200")
	appTest.run(t, userUpdate401(sd), "user-update-401")
	appTest.run(t, userUpdate400(sd), "user-update-400")

	appTest.run(t, userDelete200(sd), "user-delete-200")
	appTest.run(t, userDelete401(sd), "user-delete-401")
}
