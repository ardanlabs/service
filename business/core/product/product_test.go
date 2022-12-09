package product_test

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/product/stores/productdb"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/foundation/docker"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

var c *docker.Container

func TestMain(m *testing.M) {
	var err error
	c, err = dbtest.StartDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dbtest.StopDB(c)

	m.Run()
}

func Test_Product(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, c, "testprod")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		teardown()
	}()

	core := product.NewCore(productdb.NewStore(log, db))

	t.Log("Given the need to work with Product records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single Product.", testID)
		{
			ctx := context.Background()

			np := product.NewProduct{
				Name:     "Comic Books",
				Cost:     10,
				Quantity: 55,
				UserID:   uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
			}

			prd, err := core.Create(ctx, np)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create a product : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create a product.", dbtest.Success, testID)

			saved, err := core.QueryByID(ctx, prd.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve product by ID: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve product by ID.", dbtest.Success, testID)

			if prd.DateCreated.UnixMilli() != saved.DateCreated.UnixMilli() {
				t.Logf("\t\tTest %d:\tGot: %v", testID, saved.DateCreated)
				t.Logf("\t\tTest %d:\tExp: %v", testID, prd.DateCreated)
				t.Logf("\t\tTest %d:\tDiff: %v", testID, saved.DateCreated.Sub(prd.DateCreated))
				t.Fatalf("\t%s\tTest %d:\tShould get back the same date created.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same date created.", dbtest.Success, testID)

			if prd.DateUpdated.UnixMilli() != saved.DateUpdated.UnixMilli() {
				t.Logf("\t\tTest %d:\tGot: %v", testID, saved.DateUpdated)
				t.Logf("\t\tTest %d:\tExp: %v", testID, prd.DateUpdated)
				t.Logf("\t\tTest %d:\tDiff: %v", testID, saved.DateUpdated.Sub(prd.DateUpdated))
				t.Fatalf("\t%s\tTest %d:\tShould get back the same date updated.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same date updated.", dbtest.Success, testID)

			prd.DateCreated = time.Time{}
			prd.DateUpdated = time.Time{}
			saved.DateCreated = time.Time{}
			saved.DateUpdated = time.Time{}

			if diff := cmp.Diff(prd, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same product. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same product.", dbtest.Success, testID)

			upd := product.UpdateProduct{
				Name:     dbtest.StringPointer("Comics"),
				Cost:     dbtest.IntPointer(50),
				Quantity: dbtest.IntPointer(40),
			}

			if _, err := core.Update(ctx, saved, upd); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update product : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update product.", dbtest.Success, testID)

			saved, err = core.QueryByID(ctx, prd.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve updated product : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve updated product.", dbtest.Success, testID)

			diff := prd.DateUpdated.Sub(saved.DateUpdated)
			if diff > 0 {
				t.Fatalf("Should have a larger DateUpdated : sav %v, prd %v, dif %v", saved.DateUpdated, saved.DateUpdated, diff)
			}

			products, err := core.Query(ctx, product.QueryFilter{}, product.DefaultOrderBy, 1, 3)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve updated product : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve updated product.", dbtest.Success, testID)

			// Check specified fields were updated. Make a copy of the original product
			// and change just the fields we expect then diff it with what was saved.

			var idx int
			for i, p := range products {
				if p.ID == saved.ID {
					idx = i
				}
			}

			products[idx].DateCreated = time.Time{}
			products[idx].DateUpdated = time.Time{}
			saved.DateCreated = time.Time{}
			saved.DateUpdated = time.Time{}

			if diff := cmp.Diff(saved, products[idx]); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same product. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same product.", dbtest.Success, testID)

			upd = product.UpdateProduct{
				Name: dbtest.StringPointer("Graphic Novels"),
			}

			if _, err := core.Update(ctx, saved, upd); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update just some fields of product : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update just some fields of product.", dbtest.Success, testID)

			saved, err = core.QueryByID(ctx, prd.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve updated product : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve updated product.", dbtest.Success, testID)

			diff = prd.DateUpdated.Sub(saved.DateUpdated)
			if diff > 0 {
				t.Fatalf("Should have a larger DateUpdated : sav %v, prd %v, dif %v", saved.DateUpdated, prd.DateUpdated, diff)
			}

			if saved.Name != *upd.Name {
				t.Fatalf("\t%s\tTest %d:\tShould be able to see updated Name field : got %q want %q.", dbtest.Failed, testID, saved.Name, *upd.Name)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updated Name field.", dbtest.Success, testID)
			}

			if err := core.Delete(ctx, saved); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to delete product : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to delete product.", dbtest.Success, testID)

			_, err = core.QueryByID(ctx, prd.ID)
			if !errors.Is(err, product.ErrNotFound) {
				t.Fatalf("\t%s\tTest %d:\tShould NOT be able to retrieve deleted product : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould NOT be able to retrieve deleted product.", dbtest.Success, testID)
		}
	}
}

func Test_PagingProduct(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, c, "testpaging")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		teardown()
	}()

	core := product.NewCore(productdb.NewStore(log, db))

	t.Log("Given the need to page through product records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen paging through 2 products.", testID)
		{
			ctx := context.Background()

			name := "Comic Books"
			prd1, err := core.Query(ctx, product.QueryFilter{Name: &name}, product.DefaultOrderBy, 1, 1)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve products %q : %s.", dbtest.Failed, testID, name, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve products %q.", dbtest.Success, testID, name)

			if len(prd1) != 1 && prd1[0].Name == name {
				t.Fatalf("\t%s\tTest %d:\tShould have a single products for %q : %s.", dbtest.Failed, testID, name, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single products.", dbtest.Success, testID)

			name = "McDonalds Toys"
			prd2, err := core.Query(ctx, product.QueryFilter{Name: &name}, product.DefaultOrderBy, 1, 1)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve products %q : %s.", dbtest.Failed, testID, name, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve products %q.", dbtest.Success, testID, name)

			if len(prd2) != 1 && prd2[0].Name == name {
				t.Fatalf("\t%s\tTest %d:\tShould have a single products for %q : %s.", dbtest.Failed, testID, name, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single products.", dbtest.Success, testID)

			prd3, err := core.Query(ctx, product.QueryFilter{}, product.DefaultOrderBy, 1, 2)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve 2 products for page 1 : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve 2 products for page 1.", dbtest.Success, testID)

			if len(prd3) != 2 {
				t.Logf("\t\tTest %d:\tgot: %v", testID, len(prd3))
				t.Logf("\t\tTest %d:\texp: %v", testID, 2)
				t.Fatalf("\t%s\tTest %d:\tShould have 2 products for page 1 : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have 2 products for page 1.", dbtest.Success, testID)

			if prd3[0].ID == prd3[1].ID {
				t.Logf("\t\tTest %d:\tproduct1: %v", testID, prd3[0].ID)
				t.Logf("\t\tTest %d:\tproduct2: %v", testID, prd3[1].ID)
				t.Fatalf("\t%s\tTest %d:\tShould have different products : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have different products.", dbtest.Success, testID)
		}
	}
}
