package tests

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/data/sqldb"
	"github.com/ardanlabs/service/business/data/transaction"
	"github.com/google/go-cmp/cmp"
)

func Test_Product(t *testing.T) {
	t.Run("crud", productCrud)
	t.Run("paging", productPaging)
	t.Run("transaction", productTran)
}

func productCrud(t *testing.T) {
	seed := func(ctx context.Context, userCore *user.Core, productCore *product.Core) ([]product.Product, error) {
		var filter user.QueryFilter
		filter.WithName("Admin Gopher")

		nu1 := user.NewUser{
			Name:            "Bill Kennedy",
			Email:           mail.Address{Address: "bill@ardanlabs.com"},
			Roles:           []user.Role{user.RoleAdmin},
			Department:      "IT",
			Password:        "12345",
			PasswordConfirm: "12345",
		}
		usr1, err := userCore.Create(ctx, nu1)
		if err != nil {
			return nil, fmt.Errorf("seeding user 1 : %w", err)
		}

		prds, err := product.TestGenerateSeedProducts(1, productCore, usr1.ID)
		if err != nil {
			return nil, fmt.Errorf("seeding products : %w", err)
		}

		return prds, nil
	}

	// -------------------------------------------------------------------------

	dbTest := dbtest.NewTest(t, c, "Test_Product/crud")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		dbTest.Teardown()
	}()

	api := dbTest.Core

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Log("Go seeding ...")

	prds, err := seed(ctx, api.Crud.User, api.Crud.Product)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	saved, err := api.Crud.Product.QueryByID(ctx, prds[0].ID)
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

	if _, err := api.Crud.Product.Update(ctx, saved, upd); err != nil {
		t.Errorf("Should be able to update product : %s", err)
	}

	saved, err = api.Crud.Product.QueryByID(ctx, prds[0].ID)
	if err != nil {
		t.Fatalf("Should be able to retrieve updated product : %s", err)
	}

	diff := prds[0].DateUpdated.Sub(saved.DateUpdated)
	if diff > 0 {
		t.Fatalf("Should have a larger DateUpdated : sav %v, prd %v, dif %v", saved.DateUpdated, saved.DateUpdated, diff)
	}

	products, err := api.Crud.Product.Query(ctx, product.QueryFilter{}, product.DefaultOrderBy, 1, 3)
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

	if _, err := api.Crud.Product.Update(ctx, saved, upd); err != nil {
		t.Fatalf("Should be able to update just some fields of product : %s", err)
	}

	saved, err = api.Crud.Product.QueryByID(ctx, prds[0].ID)
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

	if err := api.Crud.Product.Delete(ctx, saved); err != nil {
		t.Fatalf("Should be able to delete product : %s", err)
	}

	_, err = api.Crud.Product.QueryByID(ctx, prds[0].ID)
	if !errors.Is(err, product.ErrNotFound) {
		t.Fatalf("Should NOT be able to retrieve deleted product : %s", err)
	}
}

func productPaging(t *testing.T) {
	seed := func(ctx context.Context, userCore *user.Core, productCore *product.Core) ([]product.Product, error) {
		var filter user.QueryFilter
		filter.WithName("Admin Gopher")

		nu1 := user.NewUser{
			Name:            "Bill Kennedy",
			Email:           mail.Address{Address: "bill@ardanlabs.com"},
			Roles:           []user.Role{user.RoleAdmin},
			Department:      "IT",
			Password:        "12345",
			PasswordConfirm: "12345",
		}
		usr1, err := userCore.Create(ctx, nu1)
		if err != nil {
			return nil, fmt.Errorf("seeding user 1 : %w", err)
		}

		prds, err := product.TestGenerateSeedProducts(2, productCore, usr1.ID)
		if err != nil {
			return nil, fmt.Errorf("seeding products : %w", err)
		}

		return prds, nil
	}

	// -------------------------------------------------------------------------

	dbTest := dbtest.NewTest(t, c, "Test_Product/paging")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		dbTest.Teardown()
	}()

	api := dbTest.Core

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Log("Go seeding ...")

	prds, err := seed(ctx, api.Crud.User, api.Crud.Product)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	name := prds[0].Name
	prd1, err := api.Crud.Product.Query(ctx, product.QueryFilter{Name: &name}, product.DefaultOrderBy, 1, 1)
	if err != nil {
		t.Fatalf("Should be able to retrieve products %q : %s", name, err)
	}

	n, err := api.Crud.Product.Count(ctx, product.QueryFilter{Name: &name})
	if err != nil {
		t.Fatalf("Should be able to retrieve product count %q : %s", name, err)
	}

	if len(prd1) == 0 || (len(prd1) != n && prd1[0].Name == name) {
		t.Log("got:", len(prd1))
		t.Log("exp:", n)
		t.Fatalf("Should have a single product for %q", name)
	}

	name = prds[1].Name
	prd2, err := api.Crud.Product.Query(ctx, product.QueryFilter{Name: &name}, product.DefaultOrderBy, 1, 1)
	if err != nil {
		t.Fatalf("Should be able to retrieve products %q : %s", name, err)
	}

	n, err = api.Crud.Product.Count(ctx, product.QueryFilter{Name: &name})
	if err != nil {
		t.Fatalf("Should be able to retrieve product count %q : %s", name, err)
	}

	if len(prd2) == 0 || (len(prd2) != n && prd2[0].Name == name) {
		t.Log("got:", len(prd2))
		t.Log("exp:", n)
		t.Fatalf("Should have a single product for %q", name)
	}

	prd3, err := api.Crud.Product.Query(ctx, product.QueryFilter{}, product.DefaultOrderBy, 1, 2)
	if err != nil {
		t.Fatalf("Should be able to retrieve 2 products for page 1 : %s", err)
	}

	n, err = api.Crud.Product.Count(ctx, product.QueryFilter{})
	if err != nil {
		t.Fatalf("Should be able to retrieve product count %q : %s", name, err)
	}

	if len(prd3) == 0 || len(prd3) != n {
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

func productTran(t *testing.T) {
	dbTest := dbtest.NewTest(t, c, "Test_Product/tran")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		dbTest.Teardown()
	}()

	api := dbTest.Core

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// -------------------------------------------------------------------------
	// Execute under a transaction with rollback

	f := func(tx transaction.Transaction) error {
		userCore, err := api.Crud.User.ExecuteUnderTransaction(tx)
		if err != nil {
			t.Fatalf("Should be able to create new user core: %s.", err)
		}

		productCore, err := api.Crud.Product.ExecuteUnderTransaction(tx)
		if err != nil {
			t.Fatalf("Should be able to create new product core: %s.", err)
		}

		email, err := mail.ParseAddress("test@test.com")
		if err != nil {
			t.Fatalf("Should be able to parse email: %s.", err)
		}

		nu := user.NewUser{
			Name:            "test user",
			Email:           *email,
			Roles:           []user.Role{user.RoleAdmin},
			Department:      "some",
			Password:        "some",
			PasswordConfirm: "some",
		}

		usr, err := userCore.Create(ctx, nu)
		if err != nil {
			return err
		}

		np := product.NewProduct{
			UserID:   usr.ID,
			Name:     "test product",
			Cost:     -1.0,
			Quantity: 1,
		}

		_, err = productCore.Create(ctx, np)
		if err != nil {
			return err
		}

		return nil
	}

	err := transaction.ExecuteUnderTransaction(ctx, dbTest.Log, sqldb.NewBeginner(dbTest.DB), f)
	if !errors.Is(err, product.ErrInvalidCost) {
		t.Fatalf("Should NOT be able to add product : %s.", err)
	}

	// -------------------------------------------------------------------------
	// Validate rollback

	email, err := mail.ParseAddress("test@test.com")
	if err != nil {
		t.Fatalf("Should be able to parse email: %s.", err)
	}

	usr, err := api.Crud.User.QueryByEmail(ctx, *email)
	if err == nil {
		t.Fatalf("Should NOT be able to retrieve user but got: %+v.", usr)
	}
	if !errors.Is(err, user.ErrNotFound) {
		t.Fatalf("Should get ErrNotFound but got: %s.", err)
	}

	count, err := api.Crud.Product.Count(ctx, product.QueryFilter{})
	if err != nil {
		t.Fatalf("Should be able to count products: %s.", err)
	}

	if count > 0 {
		t.Fatalf("Should have no products in the DB, but have: %d.", count)
	}

	// -------------------------------------------------------------------------
	// Good transaction

	f = func(tx transaction.Transaction) error {
		userCore, err := api.Crud.User.ExecuteUnderTransaction(tx)
		if err != nil {
			t.Fatalf("Should be able to create new user core: %s.", err)
		}

		productCore, err := api.Crud.Product.ExecuteUnderTransaction(tx)
		if err != nil {
			t.Fatalf("Should be able to create new product core: %s.", err)
		}

		email, err := mail.ParseAddress("test@test.com")
		if err != nil {
			t.Fatalf("Should be able to parse email: %s.", err)
		}

		nu := user.NewUser{
			Name:            "test user",
			Email:           *email,
			Roles:           []user.Role{user.RoleAdmin},
			Department:      "some",
			Password:        "some",
			PasswordConfirm: "some",
		}

		usr, err := userCore.Create(ctx, nu)
		if err != nil {
			return err
		}

		np := product.NewProduct{
			UserID:   usr.ID,
			Name:     "test product",
			Cost:     1.0,
			Quantity: 1,
		}

		_, err = productCore.Create(ctx, np)
		if err != nil {
			return err
		}

		return nil
	}

	err = transaction.ExecuteUnderTransaction(ctx, dbTest.Log, sqldb.NewBeginner(dbTest.DB), f)
	if errors.Is(err, product.ErrInvalidCost) {
		t.Fatalf("Should be able to add product : %s.", err)
	}

	// -------------------------------------------------------------------------
	// Validate

	usr, err = api.Crud.User.QueryByEmail(ctx, *email)
	if err != nil {
		t.Fatalf("Should be able to retrieve user but got: %+v.", usr)
	}

	count, err = api.Crud.Product.Count(ctx, product.QueryFilter{})
	if err != nil {
		t.Fatalf("Should be able to count products: %s.", err)
	}

	if count == 0 {
		t.Fatal("Should have products in the DB.")
	}
}
