package user_test

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/foundation/docker"
	"github.com/google/go-cmp/cmp"
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

func Test_User(t *testing.T) {
	t.Run("crud", crud)
	t.Run("paging", paging)
}

func crud(t *testing.T) {
	seed := func(ctx context.Context, usrCore *user.Core) ([]user.User, error) {
		usrs := make([]user.User, 2)

		nu1 := user.NewUser{
			Name:            "Bill Kennedy",
			Email:           mail.Address{Address: "bill@ardanlabs.com"},
			Roles:           []user.Role{user.RoleAdmin},
			Department:      "IT",
			Password:        "12345",
			PasswordConfirm: "12345",
		}
		usr1, err := usrCore.Create(ctx, nu1)
		if err != nil {
			return nil, fmt.Errorf("seeding user 1 : %w", err)
		}

		nu2 := user.NewUser{
			Name:            "Ale Kennedy",
			Email:           mail.Address{Address: "ale@ardanlabs.com"},
			Roles:           []user.Role{user.RoleUser},
			Department:      "IT",
			Password:        "12345",
			PasswordConfirm: "12345",
		}
		usr2, err := usrCore.Create(ctx, nu2)
		if err != nil {
			return nil, fmt.Errorf("seeding user 2 : %w", err)
		}

		usrs[0] = usr1
		usrs[1] = usr2

		return usrs, nil
	}

	// -------------------------------------------------------------------------

	test := dbtest.NewTest(t, c, "Test_User/crud")
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

	usrs, err := seed(ctx, api.User)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	saved, err := api.User.QueryByID(ctx, usrs[0].ID)
	if err != nil {
		t.Fatalf("Should be able to retrieve user by ID: %s.", err)
	}

	if usrs[0].DateCreated.UnixMilli() != saved.DateCreated.UnixMilli() {
		t.Logf("got: %v", saved.DateCreated)
		t.Logf("exp: %v", usrs[0].DateCreated)
		t.Logf("dif: %v", saved.DateCreated.Sub(usrs[0].DateCreated))
		t.Errorf("Should get back the same date created")
	}

	if usrs[0].DateUpdated.UnixMilli() != saved.DateUpdated.UnixMilli() {
		t.Logf("got: %v", saved.DateUpdated)
		t.Logf("exp: %v", usrs[0].DateUpdated)
		t.Logf("dif: %v", saved.DateUpdated.Sub(usrs[0].DateUpdated))
		t.Fatalf("Should get back the same date updated")
	}

	usrs[0].DateCreated = time.Time{}
	usrs[0].DateUpdated = time.Time{}
	saved.DateCreated = time.Time{}
	saved.DateUpdated = time.Time{}

	if diff := cmp.Diff(usrs[0], saved); diff != "" {
		t.Fatalf("Should get back the same user. diff:\n%s", diff)
	}

	// -------------------------------------------------------------------------

	email, err := mail.ParseAddress("jacob@ardanlabs.com")
	if err != nil {
		t.Fatalf("Should be able to parse email: %s.", err)
	}

	upd := user.UpdateUser{
		Name:       dbtest.StringPointer("Jacob Walker"),
		Email:      email,
		Department: dbtest.StringPointer("development"),
	}

	if _, err := api.User.Update(ctx, usrs[0], upd); err != nil {
		t.Fatalf("Should be able to update user : %s.", err)
	}

	saved, err = api.User.QueryByEmail(ctx, *upd.Email)
	if err != nil {
		t.Fatalf("Should be able to retrieve user by Email : %s.", err)
	}

	diff := usrs[0].DateUpdated.Sub(saved.DateUpdated)
	if diff > 0 {
		t.Errorf("Should have a larger DateUpdated : sav %v, usr %v, dif %v", saved.DateUpdated, usrs[0].DateUpdated, diff)
	}

	if saved.Name != *upd.Name {
		t.Logf("got: %v", saved.Name)
		t.Logf("exp: %v", *upd.Name)
		t.Errorf("Should be able to see updates to Name")
	}

	if saved.Email != *upd.Email {
		t.Logf("got: %v", saved.Email)
		t.Logf("exp: %v", *upd.Email)
		t.Errorf("Should be able to see updates to Email")
	}

	if saved.Department != *upd.Department {
		t.Logf("got: %v", saved.Department)
		t.Logf("exp: %v", *upd.Department)
		t.Errorf("Should be able to see updates to Department")
	}

	// -------------------------------------------------------------------------

	if err := api.User.Delete(ctx, saved); err != nil {
		t.Fatalf("Should be able to delete user : %s.", err)
	}

	_, err = api.User.QueryByID(ctx, saved.ID)
	if !errors.Is(err, user.ErrNotFound) {
		t.Fatalf("Should NOT be able to retrieve user : %s.", err)
	}
}

func paging(t *testing.T) {
	test := dbtest.NewTest(t, c, "Test_User/paging")
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

	// -------------------------------------------------------------------------

	name := "Ale Kennedy"
	users1, err := api.User.Query(ctx, user.QueryFilter{Name: &name}, user.DefaultOrderBy, 1, 1)
	if err != nil {
		t.Fatalf("Should be able to retrieve user %q : %s.", name, err)
	}

	n, err := api.User.Count(ctx, user.QueryFilter{Name: &name})
	if err != nil {
		t.Fatalf("Should be able to retrieve user count %q : %s.", name, err)
	}

	if len(users1) != n && users1[0].Name == name {
		t.Errorf("Should have a single user for %q", name)
	}

	name = "Bill Kennedy"
	users2, err := api.User.Query(ctx, user.QueryFilter{Name: &name}, user.DefaultOrderBy, 1, 1)
	if err != nil {
		t.Fatalf("Should be able to retrieve user %q : %s.", name, err)
	}

	n, err = api.User.Count(ctx, user.QueryFilter{Name: &name})
	if err != nil {
		t.Fatalf("Should be able to retrieve user count %q : %s.", name, err)
	}

	if len(users2) != n && users2[0].Name == name {
		t.Errorf("Should have a single user for %q.", name)
	}

	users3, err := api.User.Query(ctx, user.QueryFilter{}, user.DefaultOrderBy, 1, 2)
	if err != nil {
		t.Fatalf("Should be able to retrieve 2 users for page 1 : %s.", err)
	}

	n, err = api.User.Count(ctx, user.QueryFilter{})
	if err != nil {
		t.Fatalf("Should be able to retrieve user count %q : %s.", name, err)
	}

	if len(users3) != n {
		t.Logf("got: %v", len(users3))
		t.Logf("exp: %v", n)
		t.Errorf("Should have 2 users for page 1")
	}

	if users3[0].ID == users3[1].ID {
		t.Logf("User1: %v", users3[0].ID)
		t.Logf("User2: %v", users3[1].ID)
		t.Errorf("Should have different users")
	}
}
