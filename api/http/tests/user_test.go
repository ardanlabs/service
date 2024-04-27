package tests

import (
	"runtime/debug"
	"testing"
)

func Test_User(t *testing.T) {
	t.Parallel()

	// -------------------------------------------------------------------------

	apiTest := startTest(t, "Test_User")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		apiTest.dbTest.Teardown()
	}()

	// -------------------------------------------------------------------------

	sd, err := insertUserSeed(apiTest.dbTest, apiTest.auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	apiTest.run(t, userQuery200(sd), "user-query-200")
	apiTest.run(t, userQueryByID200(sd), "user-querybyid-200")

	apiTest.run(t, userCreate200(sd), "user-create-200")
	apiTest.run(t, userCreate401(sd), "user-create-401")
	apiTest.run(t, userCreate400(sd), "user-create-400")

	apiTest.run(t, userUpdate200(sd), "user-update-200")
	apiTest.run(t, userUpdate401(sd), "user-update-401")
	apiTest.run(t, userUpdate400(sd), "user-update-400")

	apiTest.run(t, userDelete200(sd), "user-delete-200")
	apiTest.run(t, userDelete401(sd), "user-delete-401")
}
