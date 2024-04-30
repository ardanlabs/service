package tests

import (
	"runtime/debug"
	"testing"
)

func Test_Product(t *testing.T) {
	t.Parallel()

	// -------------------------------------------------------------------------

	test := startTest(t, "Test_Product")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.DB.Teardown()
	}()

	// -------------------------------------------------------------------------

	sd, err := insertProductSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	test.Run(t, productQuery200(sd), "product-query-200")
	test.Run(t, productQueryByID200(sd), "product-querybyid-200")

	test.Run(t, productCreate200(sd), "product-create-200")
	test.Run(t, productCreate401(sd), "product-create-401")
	test.Run(t, productCreate400(sd), "product-create-400")

	test.Run(t, productUpdate200(sd), "product-update-200")
	test.Run(t, productUpdate401(sd), "product-update-401")
	test.Run(t, productUpdate400(sd), "product-update-400")

	test.Run(t, productDelete200(sd), "product-delete-200")
	test.Run(t, productDelete401(sd), "product-delete-401")
}
