package tests

import (
	"runtime/debug"
	"testing"
)

func Test_Product(t *testing.T) {
	t.Parallel()

	// -------------------------------------------------------------------------

	apiTest := startTest(t, "Test_Product")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		apiTest.dbTest.Teardown()
	}()

	// -------------------------------------------------------------------------

	sd, err := insertProductSeed(apiTest.dbTest, apiTest.auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	apiTest.run(t, productQuery200(sd), "product-query-200")
	apiTest.run(t, productQueryByID200(sd), "product-querybyid-200")

	apiTest.run(t, productCreate200(sd), "product-create-200")
	apiTest.run(t, productCreate401(sd), "product-create-401")
	apiTest.run(t, productCreate400(sd), "product-create-400")

	apiTest.run(t, productUpdate200(sd), "product-update-200")
	apiTest.run(t, productUpdate401(sd), "product-update-401")
	apiTest.run(t, productUpdate400(sd), "product-update-400")

	apiTest.run(t, productDelete200(sd), "product-delete-200")
	apiTest.run(t, productDelete401(sd), "product-delete-401")
}
