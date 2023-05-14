package product_test

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/foundation/docker"
	"github.com/google/go-cmp/cmp"
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
	t.Run("crud", crud)
	t.Run("paging", paging)
}

// =============================================================================

func crud(t *testing.T) {
	seed := func(ctx context.Context, usrCore *user.Core, prdCore *product.Core) ([]product.Product, error) {
		var filter user.QueryFilter
		filter.WithName("Admin Gopher")

		usrs, err := usrCore.Query(ctx, filter, user.DefaultOrderBy, 1, 1)
		if err != nil {
			return nil, fmt.Errorf("seeding users : %w", err)
		}

		prds, err := product.TestGenerateSeedProducts(1, prdCore, usrs[0].ID)
		if err != nil {
			return nil, fmt.Errorf("seeding products : %w", err)
		}

		return prds, nil
	}

	// -------------------------------------------------------------------------

	test := dbtest.NewTest(t, c)
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.Teardown()
	}()

	api := test.CoreAPIs

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Log("Go seeding ...")

	prds, err := seed(ctx, api.User, api.Product)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	saved, err := api.Product.QueryByID(ctx, prds[0].ID)
	if err != nil {
		t.Fatalf("Should be able to retrieve product by ID: %s", err)
	}

	if prds[0].DateCreated.UnixMilli() != saved.DateCreated.UnixMilli() {
		t.Logf("got: %v", saved.DateCreated)
		t.Logf("exp: %v", prds[0].DateCreated)
		t.Logf("dif: %v", saved.DateCreated.Sub(prds[0].DateCreated))
		t.Errorf("Should get back the same date created")
	}

	if prds[0].DateUpdated.UnixMilli() != saved.DateUpdated.UnixMilli() {
		t.Logf("got: %v", saved.DateUpdated)
		t.Logf("exp: %v", prds[0].DateUpdated)
		t.Logf("dif: %v", saved.DateUpdated.Sub(prds[0].DateUpdated))
		t.Errorf("Should get back the same date updated")
	}

	prds[0].DateCreated = time.Time{}
	prds[0].DateUpdated = time.Time{}
	saved.DateCreated = time.Time{}
	saved.DateUpdated = time.Time{}

	if diff := cmp.Diff(prds[0], saved); diff != "" {
		t.Errorf("Should get back the same product, dif:\n%s", diff)
	}

	// -------------------------------------------------------------------------

	upd := product.UpdateProduct{
		Name:     dbtest.StringPointer("Comics"),
		Cost:     dbtest.FloatPointer(50),
		Quantity: dbtest.IntPointer(40),
	}

	if _, err := api.Product.Update(ctx, saved, upd); err != nil {
		t.Errorf("Should be able to update product : %s", err)
	}

	saved, err = api.Product.QueryByID(ctx, prds[0].ID)
	if err != nil {
		t.Fatalf("Should be able to retrieve updated product : %s", err)
	}

	diff := prds[0].DateUpdated.Sub(saved.DateUpdated)
	if diff > 0 {
		t.Fatalf("Should have a larger DateUpdated : sav %v, prd %v, dif %v", saved.DateUpdated, saved.DateUpdated, diff)
	}

	products, err := api.Product.Query(ctx, product.QueryFilter{}, user.DefaultOrderBy, 1, 3)
	if err != nil {
		t.Fatalf("Should be able to retrieve updated product : %s", err)
	}

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
		t.Fatalf("Should get back the same product, dif:\n%s", diff)
	}

	// -------------------------------------------------------------------------

	upd = product.UpdateProduct{
		Name: dbtest.StringPointer("Graphic Novels"),
	}

	if _, err := api.Product.Update(ctx, saved, upd); err != nil {
		t.Fatalf("Should be able to update just some fields of product : %s", err)
	}

	saved, err = api.Product.QueryByID(ctx, prds[0].ID)
	if err != nil {
		t.Fatalf("Should be able to retrieve updated product : %s", err)
	}

	diff = prds[0].DateUpdated.Sub(saved.DateUpdated)
	if diff > 0 {
		t.Fatalf("Should have a larger DateUpdated : sav %v, prd %v, dif %v", saved.DateUpdated, prds[0].DateUpdated, diff)
	}

	if saved.Name != *upd.Name {
		t.Fatalf("Should be able to see updated Name field : got %q want %q", saved.Name, *upd.Name)
	}

	if err := api.Product.Delete(ctx, saved); err != nil {
		t.Fatalf("Should be able to delete product : %s", err)
	}

	_, err = api.Product.QueryByID(ctx, prds[0].ID)
	if !errors.Is(err, product.ErrNotFound) {
		t.Fatalf("Should NOT be able to retrieve deleted product : %s", err)
	}
}

func paging(t *testing.T) {
	seed := func(ctx context.Context, usrCore *user.Core, prdCore *product.Core) ([]product.Product, error) {
		var filter user.QueryFilter
		filter.WithName("Admin Gopher")

		usrs, err := usrCore.Query(ctx, filter, user.DefaultOrderBy, 1, 1)
		if err != nil {
			return nil, fmt.Errorf("seeding products : %w", err)
		}

		prds, err := product.TestGenerateSeedProducts(2, prdCore, usrs[0].ID)
		if err != nil {
			return nil, fmt.Errorf("seeding products : %w", err)
		}

		return prds, nil
	}

	// -------------------------------------------------------------------------

	test := dbtest.NewTest(t, c)
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.Teardown()
	}()

	api := test.CoreAPIs

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Log("Go seeding ...")

	prds, err := seed(ctx, api.User, api.Product)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	name := prds[0].Name
	prd1, err := api.Product.Query(ctx, product.QueryFilter{Name: &name}, user.DefaultOrderBy, 1, 1)
	if err != nil {
		t.Fatalf("Should be able to retrieve products %q : %s", name, err)
	}

	n, err := api.Product.Count(ctx, product.QueryFilter{Name: &name})
	if err != nil {
		t.Fatalf("Should be able to retrieve product count %q : %s", name, err)
	}

	if len(prd1) != n && prd1[0].Name == name {
		t.Log("got:", len(prd1))
		t.Log("exp:", n)
		t.Fatalf("Should have a single product for %q", name)
	}

	name = prds[1].Name
	prd2, err := api.Product.Query(ctx, product.QueryFilter{Name: &name}, user.DefaultOrderBy, 1, 1)
	if err != nil {
		t.Fatalf("Should be able to retrieve products %q : %s", name, err)
	}

	n, err = api.Product.Count(ctx, product.QueryFilter{Name: &name})
	if err != nil {
		t.Fatalf("Should be able to retrieve product count %q : %s", name, err)
	}

	if len(prd2) != n && prd2[0].Name == name {
		t.Log("got:", len(prd2))
		t.Log("exp:", n)
		t.Fatalf("Should have a single product for %q", name)
	}

	prd3, err := api.Product.Query(ctx, product.QueryFilter{}, user.DefaultOrderBy, 1, 2)
	if err != nil {
		t.Fatalf("Should be able to retrieve 2 products for page 1 : %s", err)
	}

	n, err = api.Product.Count(ctx, product.QueryFilter{})
	if err != nil {
		t.Fatalf("Should be able to retrieve product count %q : %s", name, err)
	}

	if len(prd3) != n {
		t.Logf("got: %v", len(prd3))
		t.Logf("exp: %v", n)
		t.Fatalf("Should have 2 products for page ")
	}

	if prd3[0].ID == prd3[1].ID {
		t.Logf("product1: %v", prd3[0].ID)
		t.Logf("product2: %v", prd3[1].ID)
		t.Fatalf("Should have different product")
	}
}
