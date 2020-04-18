package product_test

import (
	"context"
	"testing"
	"time"

	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/product"
	"github.com/ardanlabs/service/internal/tests"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

// TestProduct validates the full set of CRUD operations on Product values.
func TestProduct(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to work with Product records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single Product.", testID)
		{
			np := product.NewProduct{
				Name:     "Comic Books",
				Cost:     10,
				Quantity: 55,
			}
			now := time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)
			ctx := context.Background()

			claims := auth.NewClaims(
				"718ffbea-f4a1-4667-8ae3-b349da52675e", // This is just some random UUID.
				[]string{auth.RoleAdmin, auth.RoleUser},
				now, time.Hour,
			)

			p, err := product.Create(ctx, db, claims, np, now)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create a product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create a product.", tests.Success, testID)

			saved, err := product.Retrieve(ctx, db, p.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve product by ID: %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve product by ID.", tests.Success, testID)

			if diff := cmp.Diff(p, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same product. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same product.", tests.Success, testID)

			upd := product.UpdateProduct{
				Name:     tests.StringPointer("Comics"),
				Cost:     tests.IntPointer(50),
				Quantity: tests.IntPointer(40),
			}
			updatedTime := time.Date(2019, time.January, 1, 1, 1, 1, 0, time.UTC)

			if err := product.Update(ctx, db, claims, p.ID, upd, updatedTime); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update product.", tests.Success, testID)

			saved, err = product.Retrieve(ctx, db, p.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve updated product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve updated product.", tests.Success, testID)

			// Check specified fields were updated. Make a copy of the original product
			// and change just the fields we expect then diff it with what was saved.
			want := *p
			want.Name = *upd.Name
			want.Cost = *upd.Cost
			want.Quantity = *upd.Quantity
			want.DateUpdated = updatedTime

			if diff := cmp.Diff(want, *saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same product. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same product.", tests.Success, testID)

			upd = product.UpdateProduct{
				Name: tests.StringPointer("Graphic Novels"),
			}

			if err := product.Update(ctx, db, claims, p.ID, upd, updatedTime); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update just some fields of product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update just some fields of product.", tests.Success, testID)

			saved, err = product.Retrieve(ctx, db, p.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve updated product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve updated product.", tests.Success, testID)

			if saved.Name != *upd.Name {
				t.Fatalf("\t%s\tTest %d:\tShould be able to see updated Name field : got %q want %q.", tests.Failed, testID, saved.Name, *upd.Name)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updated Name field.", tests.Success, testID)
			}

			if err := product.Delete(ctx, db, p.ID); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to delete product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to delete product.", tests.Success, testID)

			saved, err = product.Retrieve(ctx, db, p.ID)
			if errors.Cause(err) != product.ErrNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould NOT be able to retrieve deleted product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould NOT be able to retrieve deleted product.", tests.Success, testID)
		}
	}
}
