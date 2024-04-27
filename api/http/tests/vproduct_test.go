package tests

import (
	"runtime/debug"
	"testing"
)

func Test_VProduct(t *testing.T) {
	t.Parallel()

	// -------------------------------------------------------------------------

	apiTest := startTest(t, "Test_VProduct")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		apiTest.dbTest.Teardown()
	}()

	// -------------------------------------------------------------------------

	sd, err := insertVProductSeed(apiTest.dbTest, apiTest.auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	apiTest.run(t, vproductQuery200(sd), "vproduct-query-200")
}
