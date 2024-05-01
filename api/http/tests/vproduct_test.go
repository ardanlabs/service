package tests

import (
	"runtime/debug"
	"testing"
)

func Test_VProduct(t *testing.T) {
	t.Parallel()

	// -------------------------------------------------------------------------

	test := startTest(t, "Test_VProduct")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.DB.Teardown()
	}()

	// -------------------------------------------------------------------------

	sd, err := insertVProductSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	test.Run(t, vproductQuery200(sd), "vproduct-query-200")
}
