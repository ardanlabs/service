package foobar_test

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/core/foobar"
	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/sys/core"
	"github.com/ardanlabs/service/foundation/docker"
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

func Test_Foobar(t *testing.T) {
	t.Run("transaction", transaction)
}

func transaction(t *testing.T) {
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

	np := product.NewProduct{
		Name:     "test product",
		Cost:     -1.0,
		Quantity: 1,
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

	tx, err := test.DB.Beginx()
	if err != nil {
		t.Fatalf("Should NOT to begin transaction: %s.", err)
	}

	tran := func(c *foobar.Core) error {
		_, err = c.Create(ctx, np, nu)
		return err
	}
	err = core.WithinTranCore[*foobar.Core](ctx, test.Log, tx, api.Foobar, tran)
	if !errors.Is(err, product.ErrInvalidCost) {
		t.Fatalf("Should NOT be able to add product : %s.", err)
	}

	usr, err := api.User.QueryByEmail(ctx, nu.Email)
	if err == nil {
		t.Fatalf("Should NOT be able to retrieve user but got: %+v.", usr)
	} else {
		if !errors.Is(err, user.ErrNotFound) {
			t.Fatalf("Should got ErrNotFound but got: %s.", err)
		}
	}
	count, err := api.Product.Count(ctx, product.QueryFilter{})
	if err != nil {
		t.Fatalf("Should be able to count products: %s.", err)
	}
	if count > 0 {
		t.Fatalf("Should have no products in the DB, but have: %d.", count)
	}

	// -------------------------------------------------------------------------

	np.Cost = 4.0
	prd, err := api.Foobar.Create(ctx, np, nu)
	if err != nil {
		t.Fatalf("Should be able to create user and product: %s.", err)
	}

	_, err = api.User.QueryByEmail(ctx, nu.Email)
	if err != nil {
		t.Fatalf("Should have user in db: %+s.", err)
	}
	_, err = api.Product.QueryByID(ctx, prd.ID)
	if err != nil {
		t.Fatalf("Should have product in db: %+s.", err)
	}
}
