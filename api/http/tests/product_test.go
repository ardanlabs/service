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
		apiTest.DB.Teardown()
	}()

	// -------------------------------------------------------------------------

	sd, err := insertProductSeed(apiTest.DB, apiTest.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	apiTest.Run(t, productQuery200(sd), "product-query-200")
	apiTest.Run(t, productQueryByID200(sd), "product-querybyid-200")

	apiTest.Run(t, productCreate200(sd), "product-create-200")
	apiTest.Run(t, productCreate401(sd), "product-create-401")
	apiTest.Run(t, productCreate400(sd), "product-create-400")

	apiTest.Run(t, productUpdate200(sd), "product-update-200")
	apiTest.Run(t, productUpdate401(sd), "product-update-401")
	apiTest.Run(t, productUpdate400(sd), "product-update-400")

	apiTest.Run(t, productDelete200(sd), "product-delete-200")
	apiTest.Run(t, productDelete401(sd), "product-delete-401")
}
