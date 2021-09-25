package product_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/data/dbtest"
	store "github.com/ardanlabs/service/business/data/store/product"
	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/ardanlabs/service/business/sys/validate"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-cmp/cmp"
)

var dbc = dbtest.DBContainer{
	Image: "postgres:13-alpine",
	Port:  "5432",
	Args:  []string{"-e", "POSTGRES_PASSWORD=postgres"},
}

func TestProduct(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, dbc)
	t.Cleanup(teardown)

	product := product.NewCore(log, db)

	t.Log("Given the need to work with Product records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single Product.", testID)
		{
			ctx := context.Background()
			now := time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)

			claims := auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer:    "service project",
					Subject:   "5cf37266-3473-4006-984f-9325122678b7",
					ExpiresAt: time.Now().Add(time.Hour).Unix(),
					IssuedAt:  time.Now().UTC().Unix(),
				},
				Roles: []string{auth.RoleAdmin, auth.RoleUser},
			}

			np := store.NewProduct{
				Name:     "Comic Books",
				Cost:     10,
				Quantity: 55,
				UserID:   "5cf37266-3473-4006-984f-9325122678b7",
			}

			prd, err := product.Create(ctx, np, now)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create a product : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create a product.", dbtest.Success, testID)

			saved, err := product.QueryByID(ctx, prd.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve product by ID: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve product by ID.", dbtest.Success, testID)

			if diff := cmp.Diff(prd, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same product. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same product.", dbtest.Success, testID)

			upd := store.UpdateProduct{
				Name:     dbtest.StringPointer("Comics"),
				Cost:     dbtest.IntPointer(50),
				Quantity: dbtest.IntPointer(40),
			}
			updatedTime := time.Date(2019, time.January, 1, 1, 1, 1, 0, time.UTC)

			if err := product.Update(ctx, claims, prd.ID, upd, updatedTime); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update product : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update product.", dbtest.Success, testID)

			products, err := product.Query(ctx, 1, 3)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve updated product : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve updated product.", dbtest.Success, testID)

			// Check specified fields were updated. Make a copy of the original product
			// and change just the fields we expect then diff it with what was saved.
			want := prd
			want.Name = *upd.Name
			want.Cost = *upd.Cost
			want.Quantity = *upd.Quantity
			want.DateUpdated = updatedTime

			var idx int
			for i, p := range products {
				if p.ID == want.ID {
					idx = i
				}
			}
			if diff := cmp.Diff(want, products[idx]); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same product. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same product.", dbtest.Success, testID)

			upd = store.UpdateProduct{
				Name: dbtest.StringPointer("Graphic Novels"),
			}

			if err := product.Update(ctx, claims, prd.ID, upd, updatedTime); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update just some fields of product : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update just some fields of product.", dbtest.Success, testID)

			saved, err = product.QueryByID(ctx, prd.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve updated product : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve updated product.", dbtest.Success, testID)

			if saved.Name != *upd.Name {
				t.Fatalf("\t%s\tTest %d:\tShould be able to see updated Name field : got %q want %q.", dbtest.Failed, testID, saved.Name, *upd.Name)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updated Name field.", dbtest.Success, testID)
			}

			if err := product.Delete(ctx, claims, prd.ID); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to delete product : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to delete product.", dbtest.Success, testID)

			_, err = product.QueryByID(ctx, prd.ID)
			if !errors.Is(err, validate.ErrNotFound) {
				t.Fatalf("\t%s\tTest %d:\tShould NOT be able to retrieve deleted product : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould NOT be able to retrieve deleted product.", dbtest.Success, testID)
		}
	}
}
