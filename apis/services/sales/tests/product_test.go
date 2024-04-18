package tests

import (
	"runtime/debug"
	"testing"
)

func Test_Product(t *testing.T) {
	t.Parallel()

	// -------------------------------------------------------------------------

	dbTest, appTest := startTest(t, "Test_Product")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		dbTest.Teardown()
	}()

	// -------------------------------------------------------------------------

	sd, err := insertProductSeed(dbTest)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	appTest.run(t, productQuery200(sd), "product-query-200")
	appTest.run(t, productQueryByID200(sd), "product-querybyid-200")

	appTest.run(t, productCreate200(sd), "product-create-200")
	appTest.run(t, productCreate401(sd), "product-create-401")
	appTest.run(t, productCreate400(sd), "product-create-400")

	appTest.run(t, productUpdate200(sd), "product-update-200")
	appTest.run(t, productUpdate401(sd), "product-update-401")
	appTest.run(t, productUpdate400(sd), "product-update-400")

	appTest.run(t, productDelete200(sd), "product-delete-200")
	appTest.run(t, productDelete401(sd), "product-delete-401")
}
