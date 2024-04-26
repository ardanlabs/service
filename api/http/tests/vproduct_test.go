package tests

import (
	"runtime/debug"
	"testing"
)

func Test_VProduct(t *testing.T) {
	t.Parallel()

	// -------------------------------------------------------------------------

	appTest := startTest(t, "Test_VProduct")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		appTest.dbTest.Teardown()
	}()

	// -------------------------------------------------------------------------

	sd, err := insertVProductSeed(appTest.dbTest, appTest.auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	appTest.run(t, vproductQuery200(sd), "vproduct-query-200")
}
