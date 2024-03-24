package tests

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/google/go-cmp/cmp"
)

func Test_User(t *testing.T) {
	t.Run("crud", userCrud)
	t.Run("paging", userPaging)
}

func userCrud(t *testing.T) {
	seed := func(ctx context.Context, userCore *user.Core) ([]user.User, error) {
		usrs, err := user.TestGenerateSeedUsers(ctx, 2, user.RoleAdmin, userCore)
		if err != nil {
			return nil, fmt.Errorf("seeding user : %w", err)
		}

		return usrs, nil
	}

	// -------------------------------------------------------------------------

	dbTest := dbtest.NewTest(t, c, "Test_User/crud")
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

	usrs, err := seed(ctx, api.Crud.User)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	saved, err := api.Crud.User.QueryByID(ctx, usrs[0].ID)
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
		Roles:      []user.Role{user.RoleUser},
		Enabled:    dbtest.BoolPointer(false),
	}

	if _, err := api.Crud.User.Update(ctx, usrs[0], upd); err != nil {
		t.Fatalf("Should be able to update user : %s.", err)
	}

	saved, err = api.Crud.User.QueryByEmail(ctx, *upd.Email)
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

	if len(saved.Roles) != len(upd.Roles) {
		t.Logf("got: %v", saved.Roles)
		t.Logf("exp: %v", upd.Roles)
		t.Errorf("Should be able to see updates to Roles")
	} else {
		for i := range saved.Roles {
			if saved.Roles[i] != upd.Roles[i] {
				t.Logf("got: %v", saved.Roles)
				t.Logf("exp: %v", upd.Roles)
				t.Errorf("Should be able to see updates to Roles")
				break
			}
		}
	}

	if saved.Enabled != *upd.Enabled {
		t.Logf("got: %v", saved.Enabled)
		t.Logf("exp: %v", *upd.Enabled)
		t.Errorf("Should be able to see updates to Enabled")
	}

	// -------------------------------------------------------------------------

	if err := api.Crud.User.Delete(ctx, saved); err != nil {
		t.Fatalf("Should be able to delete user : %s.", err)
	}

	_, err = api.Crud.User.QueryByID(ctx, saved.ID)
	if !errors.Is(err, user.ErrNotFound) {
		t.Fatalf("Should NOT be able to retrieve user : %s.", err)
	}
}

func userPaging(t *testing.T) {
	seed := func(ctx context.Context, userCore *user.Core) ([]user.User, error) {
		usrs := make([]user.User, 2)

		usrsAdmin, err := user.TestGenerateSeedUsers(ctx, 1, user.RoleAdmin, userCore)
		if err != nil {
			return nil, fmt.Errorf("seeding user : %w", err)
		}

		usrsUser, err := user.TestGenerateSeedUsers(ctx, 1, user.RoleUser, userCore)
		if err != nil {
			return nil, fmt.Errorf("seeding user : %w", err)
		}

		usrs[0] = usrsAdmin[0]
		usrs[1] = usrsUser[0]

		return usrs, nil
	}

	// -------------------------------------------------------------------------

	dbTest := dbtest.NewTest(t, c, "Test_User/paging")
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

	usrs, err := seed(ctx, api.Crud.User)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	name := usrs[0].Name
	users1, err := api.Crud.User.Query(ctx, user.QueryFilter{Name: &name}, user.DefaultOrderBy, 1, 1)
	if err != nil {
		t.Fatalf("Should be able to retrieve user %q : %s.", name, err)
	}

	n, err := api.Crud.User.Count(ctx, user.QueryFilter{Name: &name})
	if err != nil {
		t.Fatalf("Should be able to retrieve user count %q : %s.", name, err)
	}

	if len(users1) == 0 || (len(users1) != n && users1[0].Name == name) {
		t.Errorf("Should have a single user for %q", name)
	}

	name = usrs[1].Name
	users2, err := api.Crud.User.Query(ctx, user.QueryFilter{Name: &name}, user.DefaultOrderBy, 1, 1)
	if err != nil {
		t.Fatalf("Should be able to retrieve user %q : %s.", name, err)
	}

	n, err = api.Crud.User.Count(ctx, user.QueryFilter{Name: &name})
	if err != nil {
		t.Fatalf("Should be able to retrieve user count %q : %s.", name, err)
	}

	if len(users2) == 0 || (len(users2) != n && users2[0].Name == name) {
		t.Errorf("Should have a single user for %q.", name)
	}

	users3, err := api.Crud.User.Query(ctx, user.QueryFilter{}, user.DefaultOrderBy, 1, 4)
	if err != nil {
		t.Fatalf("Should be able to retrieve 2 users for page 1 : %s.", err)
	}

	n, err = api.Crud.User.Count(ctx, user.QueryFilter{})
	if err != nil {
		t.Fatalf("Should be able to retrieve user count %q : %s.", name, err)
	}

	if len(users3) == 0 || len(users3) != n {
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
