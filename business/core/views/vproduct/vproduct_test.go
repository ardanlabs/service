package vproduct_test

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/core/views/vproduct"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/foundation/docker"
)

var c *docker.Container

func TestMain(m *testing.M) {
	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}

	os.Exit(code)
}

func run(m *testing.M) (int, error) {
	var err error

	c, err = dbtest.StartDB()
	if err != nil {
		return 1, err
	}
	defer dbtest.StopDB(c)

	return m.Run(), nil
}

func Test_VProduct(t *testing.T) {
	t.Run("paging", paging)
}

func paging(t *testing.T) {
	seed := func(ctx context.Context, usrCore *user.Core, prdCore *product.Core) ([]product.Product, []user.User, error) {
		var filter user.QueryFilter
		filter.WithName("Admin Gopher")

		usrs, err := usrCore.Query(ctx, filter, user.DefaultOrderBy, 1, 1)
		if err != nil {
			return nil, nil, fmt.Errorf("seeding products : %w", err)
		}

		prds, err := product.TestGenerateSeedProducts(2, prdCore, usrs[0].ID)
		if err != nil {
			return nil, nil, fmt.Errorf("seeding products : %w", err)
		}

		return prds, usrs, nil
	}

	// -------------------------------------------------------------------------

	test := dbtest.NewTest(t, c, "Test_VProduct/paging")
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

	prds, usrs, err := seed(ctx, api.User, api.Product)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	name := prds[0].Name
	prd1, err := api.VProduct.Query(ctx, vproduct.QueryFilter{Name: &name}, vproduct.DefaultOrderBy, 1, 1)
	if err != nil {
		t.Fatalf("Should be able to retrieve products %q : %s", name, err)
	}

	n, err := api.VProduct.Count(ctx, vproduct.QueryFilter{Name: &name})
	if err != nil {
		t.Fatalf("Should be able to retrieve product count %q : %s", name, err)
	}

	if len(prd1) != n && prd1[0].Name == name {
		t.Log("got:", len(prd1))
		t.Log("exp:", n)
		t.Fatalf("Should have a single product for %q", name)
	}

	if prd1[0].UserName != usrs[0].Name {
		t.Log("got:", prd1[0].UserName)
		t.Log("exp:", usrs[0].Name)
		t.Fatal("Should have the correct user name")
	}

	name = prds[1].Name
	prd2, err := api.VProduct.Query(ctx, vproduct.QueryFilter{Name: &name}, vproduct.DefaultOrderBy, 1, 1)
	if err != nil {
		t.Fatalf("Should be able to retrieve products %q : %s", name, err)
	}

	n, err = api.VProduct.Count(ctx, vproduct.QueryFilter{Name: &name})
	if err != nil {
		t.Fatalf("Should be able to retrieve product count %q : %s", name, err)
	}

	if len(prd2) != n && prd2[0].Name == name {
		t.Log("got:", len(prd2))
		t.Log("exp:", n)
		t.Fatalf("Should have a single product for %q", name)
	}

	if prd2[0].UserName != usrs[0].Name {
		t.Log("got:", prd2[0].UserName)
		t.Log("exp:", usrs[0].Name)
		t.Fatal("Should have the correct user name")
	}

	prd3, err := api.VProduct.Query(ctx, vproduct.QueryFilter{}, vproduct.DefaultOrderBy, 1, 2)
	if err != nil {
		t.Fatalf("Should be able to retrieve 2 products for page 1 : %s", err)
	}

	n, err = api.VProduct.Count(ctx, vproduct.QueryFilter{})
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

	if prd3[0].UserName != usrs[0].Name {
		t.Log("got:", prd3[0].UserName)
		t.Log("exp:", usrs[0].Name)
		t.Fatal("Should have the correct user name")
	}
}
