package product_test

import (
	"os"
	"testing"
	"time"

	"github.com/ardanlabs/service/internal/platform/tests"
	"github.com/ardanlabs/service/internal/product"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

var test *tests.Test

// TestMain is the entry point for testing.
func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	test = tests.New()
	defer test.TearDown()
	return m.Run()
}

// TestCreate validates we can create a product and it exists in the DB.
func TestProduct(t *testing.T) {
	defer tests.Recover(t)

	t.Log("Given the need to work with Product records.")
	{
		t.Log("\tWhen handling a single Product.")
		{
			ctx := tests.Context()

			dbConn := test.MasterDB.Copy()
			defer dbConn.Close()

			np := product.NewProduct{
				Name:      "Comic Books",
				Notes:     "Various conditions.",
				Family:    "Kennedy",
				UnitPrice: 25,
				Quantity:  60,
			}

			p, err := product.Create(ctx, dbConn, &np, time.Now().UTC())
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create a product : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create a product.", tests.Success)

			savedP, err := product.Retrieve(ctx, dbConn, p.ID.Hex())
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retrieve product by ID: %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retrieve product by ID.", tests.Success)

			if diff := cmp.Diff(p, savedP); diff != "" {
				t.Fatalf("\t%s\tShould get back the same product. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get back the same product.", tests.Success)

			upd := product.UpdateProduct{
				Name:      tests.StringPointer("Comics"),
				Notes:     tests.StringPointer(""),
				Family:    tests.StringPointer("walker"),
				UnitPrice: tests.IntPointer(50),
				Quantity:  tests.IntPointer(40),
			}

			if err := product.Update(ctx, dbConn, p.ID.Hex(), upd, time.Now().UTC()); err != nil {
				t.Fatalf("\t%s\tShould be able to update product : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to update product.", tests.Success)

			savedP, err = product.Retrieve(ctx, dbConn, p.ID.Hex())
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retrieve updated product : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retrieve updated product.", tests.Success)

			// Build a product matching what we expect to see. We just use the
			// modified time from the database.
			want := &product.Product{
				ID:           p.ID,
				Name:         *upd.Name,
				Notes:        *upd.Notes,
				Family:       *upd.Family,
				UnitPrice:    *upd.UnitPrice,
				Quantity:     *upd.Quantity,
				DateCreated:  p.DateCreated,
				DateModified: savedP.DateModified,
			}

			if diff := cmp.Diff(want, savedP); diff != "" {
				t.Fatalf("\t%s\tShould get back the same product. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get back the same product.", tests.Success)

			upd = product.UpdateProduct{
				Name: tests.StringPointer("Graphic Novels"),
			}

			if err := product.Update(ctx, dbConn, p.ID.Hex(), upd, time.Now().UTC()); err != nil {
				t.Fatalf("\t%s\tShould be able to update just some fields of product : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to update just some fields of product.", tests.Success)

			savedP, err = product.Retrieve(ctx, dbConn, p.ID.Hex())
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retrieve updated product : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retrieve updated product.", tests.Success)

			if savedP.Name != *upd.Name {
				t.Fatalf("\t%s\tShould be able to see updated Name field : got %q want %q.", tests.Failed, savedP.Name, *upd.Name)
			} else {
				t.Logf("\t%s\tShould be able to see updated Name field.", tests.Success)
			}

			if savedP.Family != "walker" {
				t.Fatalf("\t%s\tShould not see updates in other fields : Family was %q want %q.", tests.Failed, savedP.Family, "walker")
			} else {
				t.Logf("\t%s\tShould not see updates in other fields.", tests.Success)
			}

			if err := product.Delete(ctx, dbConn, p.ID.Hex()); err != nil {
				t.Fatalf("\t%s\tShould be able to delete product : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to delete product.", tests.Success)

			savedP, err = product.Retrieve(ctx, dbConn, p.ID.Hex())
			if errors.Cause(err) != product.ErrNotFound {
				t.Fatalf("\t%s\tShould NOT be able to retrieve deleted product : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould NOT be able to retrieve deleted product.", tests.Success)
		}
	}
}
