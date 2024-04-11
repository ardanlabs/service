package tests

import (
	"runtime/debug"
	"testing"
)

func Test_VProduct(t *testing.T) {
	t.Parallel()

	// -------------------------------------------------------------------------

	dbTest, appTest := startTest(t, "Test_VProduct")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		dbTest.Teardown()
	}()

	// -------------------------------------------------------------------------

	sd, err := insertVProductSeed(dbTest)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	appTest.Run(t, vproductQuery200(sd), "vproduct-query-200")
}
