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
		apiTest.DB.Teardown()
	}()

	// -------------------------------------------------------------------------

	sd, err := insertVProductSeed(apiTest.DB, apiTest.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	apiTest.Run(t, vproductQuery200(sd), "vproduct-query-200")
}
