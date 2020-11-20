package product_test

import (
	"context"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/auth"
	"github.com/ardanlabs/service/business/data/product"
	"github.com/ardanlabs/service/business/tests"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

func TestProduct(t *testing.T) {
	log, db, teardown := tests.NewUnit(t)
	t.Cleanup(teardown)

	p := product.New(log, db)

	t.Log("Given the need to work with Product records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single Product.", testID)
		{
			ctx := context.Background()
			now := time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)
			traceID := "00000000-0000-0000-0000-000000000000"

			claims := auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer:    "service project",
					Subject:   "718ffbea-f4a1-4667-8ae3-b349da52675e",
					Audience:  "students",
					ExpiresAt: now.Add(time.Hour).Unix(),
					IssuedAt:  now.Unix(),
				},
				Roles: []string{auth.RoleAdmin, auth.RoleUser},
			}

			np := product.NewProduct{
				Name:     "Comic Books",
				Cost:     10,
				Quantity: 55,
			}

			prd, err := p.Create(ctx, traceID, claims, np, now)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create a product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create a product.", tests.Success, testID)

			saved, err := p.QueryByID(ctx, traceID, prd.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve product by ID: %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve product by ID.", tests.Success, testID)

			if diff := cmp.Diff(prd, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same product. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same product.", tests.Success, testID)

			upd := product.UpdateProduct{
				Name:     tests.StringPointer("Comics"),
				Cost:     tests.IntPointer(50),
				Quantity: tests.IntPointer(40),
			}
			updatedTime := time.Date(2019, time.January, 1, 1, 1, 1, 0, time.UTC)

			if err := p.Update(ctx, traceID, claims, prd.ID, upd, updatedTime); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update product.", tests.Success, testID)

			products, err := p.Query(ctx, traceID, 1, 1)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve updated product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve updated product.", tests.Success, testID)

			// Check specified fields were updated. Make a copy of the original product
			// and change just the fields we expect then diff it with what was saved.
			want := prd
			want.Name = *upd.Name
			want.Cost = *upd.Cost
			want.Quantity = *upd.Quantity
			want.DateUpdated = updatedTime

			if diff := cmp.Diff(want, products[0]); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same product. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same product.", tests.Success, testID)

			upd = product.UpdateProduct{
				Name: tests.StringPointer("Graphic Novels"),
			}

			if err := p.Update(ctx, traceID, claims, prd.ID, upd, updatedTime); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update just some fields of product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update just some fields of product.", tests.Success, testID)

			saved, err = p.QueryByID(ctx, traceID, prd.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve updated product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve updated product.", tests.Success, testID)

			if saved.Name != *upd.Name {
				t.Fatalf("\t%s\tTest %d:\tShould be able to see updated Name field : got %q want %q.", tests.Failed, testID, saved.Name, *upd.Name)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updated Name field.", tests.Success, testID)
			}

			if err := p.Delete(ctx, traceID, claims, prd.ID); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to delete product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to delete product.", tests.Success, testID)

			_, err = p.QueryByID(ctx, traceID, prd.ID)
			if errors.Cause(err) != product.ErrNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould NOT be able to retrieve deleted product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould NOT be able to retrieve deleted product.", tests.Success, testID)
		}
	}
}
