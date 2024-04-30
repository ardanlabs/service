package tests

import (
	"runtime/debug"
	"testing"
)

func Test_User(t *testing.T) {
	t.Parallel()

	// -------------------------------------------------------------------------

	test := startTest(t, "Test_User")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.DB.Teardown()
	}()

	// -------------------------------------------------------------------------

	sd, err := insertUserSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	test.Run(t, userQuery200(sd), "user-query-200")
	test.Run(t, userQueryByID200(sd), "user-querybyid-200")

	test.Run(t, userCreate200(sd), "user-create-200")
	test.Run(t, userCreate401(sd), "user-create-401")
	test.Run(t, userCreate400(sd), "user-create-400")

	test.Run(t, userUpdate200(sd), "user-update-200")
	test.Run(t, userUpdate401(sd), "user-update-401")
	test.Run(t, userUpdate400(sd), "user-update-400")

	test.Run(t, userDelete200(sd), "user-delete-200")
	test.Run(t, userDelete401(sd), "user-delete-401")
}
