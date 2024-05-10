package user_test

import (
	"runtime/debug"
	"testing"

	"github.com/ardanlabs/service/api/sdk/http/apitest"
)

func Test_User(t *testing.T) {
	t.Parallel()

	// -------------------------------------------------------------------------

	test := apitest.StartTest(t, "Test_User")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.DB.Teardown()
	}()

	// -------------------------------------------------------------------------

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	test.Run(t, query200(sd), "query-200")
	test.Run(t, queryByID200(sd), "querybyid-200")

	test.Run(t, create200(sd), "create-200")
	test.Run(t, create401(sd), "create-401")
	test.Run(t, create400(sd), "create-400")

	test.Run(t, update200(sd), "update-200")
	test.Run(t, update401(sd), "update-401")
	test.Run(t, update400(sd), "update-400")

	test.Run(t, delete200(sd), "delete-200")
	test.Run(t, delete401(sd), "delete-401")
}
