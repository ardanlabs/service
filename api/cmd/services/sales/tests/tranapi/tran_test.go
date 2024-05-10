package tran_test

import (
	"runtime/debug"
	"testing"

	"github.com/ardanlabs/service/api/sdk/http/apitest"
)

func Test_Tran(t *testing.T) {
	t.Parallel()

	// -------------------------------------------------------------------------

	test := apitest.StartTest(t, "Test_VProduct")
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

	test.Run(t, create200(sd), "query-200")
}
