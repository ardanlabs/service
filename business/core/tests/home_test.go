package tests

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/core/crud/home"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/google/go-cmp/cmp"
)

func Test_Home(t *testing.T) {
	t.Run("crud", homeCrud)
	t.Run("paging", homePaging)
}

func homeCrud(t *testing.T) {
	seed := func(ctx context.Context, userCore *user.Core, homeCore *home.Core) ([]home.Home, error) {
		hmes := make([]home.Home, 1)

		var filter user.QueryFilter
		filter.WithName("Admin Gopher")

		nu := user.NewUser{
			Name:            "Bill Kennedy",
			Email:           mail.Address{Address: "bill@ardanlabs.com"},
			Roles:           []user.Role{user.RoleAdmin},
			Department:      "IT",
			Password:        "12345",
			PasswordConfirm: "12345",
		}

		usr, err := userCore.Create(ctx, nu)
		if err != nil {
			return nil, fmt.Errorf("seeding user : %w", err)
		}

		nh := home.NewHome{
			UserID: usr.ID,
			Type:   home.TypeSingle,
			Address: home.Address{
				Address1: "Address1",
				Address2: "Address1",
				ZipCode:  "12345",
				City:     "City1",
				State:    "State1",
				Country:  "Country1",
			},
		}

		hme, err := homeCore.Create(ctx, nh)
		if err != nil {
			return nil, fmt.Errorf("seeding home : %w", err)
		}

		hmes[0] = hme

		return hmes, nil
	}

	// ---------------------------------------------------------------------------

	dbTest := dbtest.NewTest(t, c, "Test_Home/crud")

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

	hmes, err := seed(ctx, api.Crud.User, api.Crud.Home)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// ---------------------------------------------------------------------------

	saved, err := api.Crud.Home.QueryByID(ctx, hmes[0].ID)
	if err != nil {
		t.Fatalf("Should be able to retrieve home by ID: %s", err)
	}

	if hmes[0].DateCreated.UnixMilli() != saved.DateCreated.UnixMilli() {
		t.Logf("got: %v", saved.DateCreated)
		t.Logf("exp: %v", hmes[0].DateCreated)
		t.Logf("dif: %v", saved.DateCreated.Sub(hmes[0].DateCreated))
		t.Errorf("Should get back the same date created")
	}

	if hmes[0].DateUpdated.UnixMilli() != saved.DateUpdated.UnixMilli() {
		t.Logf("got: %v", saved.DateUpdated)
		t.Logf("exp: %v", hmes[0].DateUpdated)
		t.Logf("dif: %v", saved.DateUpdated.Sub(hmes[0].DateUpdated))
		t.Errorf("Should get back the same date updated")
	}

	hmes[0].DateCreated = time.Time{}
	hmes[0].DateUpdated = time.Time{}
	saved.DateCreated = time.Time{}
	saved.DateUpdated = time.Time{}

	if diff := cmp.Diff(hmes[0], saved); diff != "" {
		t.Errorf("Should get back the same home, dif:\n%s", diff)
	}

	// ---------------------------------------------------------------------------

	upd := home.UpdateHome{
		Address: &home.UpdateAddress{
			Address1: dbtest.StringPointer("Fake St. 123"),
			Address2: dbtest.StringPointer("Apt 6942"),
			ZipCode:  dbtest.StringPointer("443223"),
			City:     dbtest.StringPointer("Austin"),
			State:    dbtest.StringPointer("Texas"),
			Country:  dbtest.StringPointer("US"),
		},
		Type: &home.TypeSingle,
	}

	if _, err := api.Crud.Home.Update(ctx, saved, upd); err != nil {
		t.Errorf("Should be able to update home : %s", err)
	}

	saved, err = api.Crud.Home.QueryByID(ctx, hmes[0].ID)
	if err != nil {
		t.Fatalf("Should be able to retrieve updated home : %s", err)
	}

	diff := hmes[0].DateUpdated.Sub(saved.DateUpdated)
	if diff > 0 {
		t.Fatalf("Should have a larger DateUpdated : sav %v, hme %v, dif %v", saved.DateUpdated, saved.DateUpdated, diff)
	}

	homes, err := api.Crud.Home.Query(ctx, home.QueryFilter{}, user.DefaultOrderBy, 1, 3)
	if err != nil {
		t.Fatalf("Should be able to retrieve updated home : %s", err)
	}

	// Check specified fields were updated. Make a copy of the original home
	// and change just the fields we expect then diff it with what was saved.

	var idx int
	for i, h := range homes {
		if h.ID == saved.ID {
			idx = i
		}
	}

	homes[idx].DateCreated = time.Time{}
	homes[idx].DateUpdated = time.Time{}
	saved.DateCreated = time.Time{}
	saved.DateUpdated = time.Time{}

	if diff := cmp.Diff(saved, homes[idx]); diff != "" {
		t.Fatalf("Should get back the same home, dif:\n%s", diff)
	}

	// -------------------------------------------------------------------------

	upd = home.UpdateHome{
		Type: &home.TypeCondo,
	}

	if _, err := api.Crud.Home.Update(ctx, saved, upd); err != nil {
		t.Fatalf("Should be able to update just some fields of home : %s", err)
	}

	saved, err = api.Crud.Home.QueryByID(ctx, hmes[0].ID)
	if err != nil {
		t.Fatalf("Should be able to retrieve updated home : %s", err)
	}

	diff = hmes[0].DateUpdated.Sub(saved.DateUpdated)
	if diff > 0 {
		t.Fatalf("Should have a larger DateUpdated : sav %v, hme %v, dif %v", saved.DateUpdated, hmes[0].DateUpdated, diff)
	}

	if saved.Type != *upd.Type {
		t.Fatalf("Should be able to see updated Type field : got %q want %q", saved.Type, *upd.Type)
	}

	if err := api.Crud.Home.Delete(ctx, saved); err != nil {
		t.Fatalf("Should be able to delete home : %s", err)
	}

	_, err = api.Crud.Home.QueryByID(ctx, hmes[0].ID)
	if !errors.Is(err, home.ErrNotFound) {
		t.Fatalf("Should NOT be able to retrieve deleted home : %s", err)
	}
}

func homePaging(t *testing.T) {
	seed := func(ctx context.Context, userCore *user.Core, homeCore *home.Core) ([]home.Home, error) {
		hmes := make([]home.Home, 2)

		var filter user.QueryFilter
		filter.WithName("Admin Gopher")

		nu := user.NewUser{
			Name:            "Bill Kennedy",
			Email:           mail.Address{Address: "bill@ardanlabs.com"},
			Roles:           []user.Role{user.RoleAdmin},
			Department:      "IT",
			Password:        "12345",
			PasswordConfirm: "12345",
		}

		usr, err := userCore.Create(ctx, nu)
		if err != nil {
			return nil, fmt.Errorf("seeding user : %w", err)
		}

		nh1 := home.NewHome{
			UserID: usr.ID,
			Type:   home.TypeSingle,
			Address: home.Address{
				Address1: "Address1",
				Address2: "Address1",
				ZipCode:  "12345",
				City:     "City1",
				State:    "State1",
				Country:  "Country1",
			},
		}

		hme1, err := homeCore.Create(ctx, nh1)
		if err != nil {
			return nil, fmt.Errorf("seeding home 1 : %w", err)
		}

		nh2 := home.NewHome{
			UserID: usr.ID,
			Type:   home.TypeSingle,
			Address: home.Address{
				Address1: "Address2",
				Address2: "Address2",
				ZipCode:  "67891",
				City:     "City2",
				State:    "State2",
				Country:  "Country2",
			},
		}

		hme2, err := homeCore.Create(ctx, nh2)
		if err != nil {
			return nil, fmt.Errorf("seeding home 2 : %w", err)
		}

		hmes[0] = hme1
		hmes[0] = hme2

		return hmes, nil
	}

	// -------------------------------------------------------------------------

	dbTest := dbtest.NewTest(t, c, "Test_Home/paging")
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

	hmes, err := seed(ctx, api.Crud.User, api.Crud.Home)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	homeType := hmes[0].Type
	hme1, err := api.Crud.Home.Query(ctx, home.QueryFilter{Type: &homeType}, user.DefaultOrderBy, 1, 2)
	if err != nil {
		t.Fatalf("Should be able to retrieve homes %q : %s", homeType, err)
	}

	n, err := api.Crud.Home.Count(ctx, home.QueryFilter{Type: &homeType})
	if err != nil {
		t.Fatalf("Should be able to retrieve home count %q : %s", homeType, err)
	}

	if len(hme1) == 0 || len(hme1) != n {
		t.Log("got:", len(hme1))
		t.Log("exp:", n)
		t.Fatal("Should have the correct number of homes")
	}

	hme2, err := api.Crud.Home.Query(ctx, home.QueryFilter{}, user.DefaultOrderBy, 1, 2)
	if err != nil {
		t.Fatalf("Should be able to retrieve 2 homes for page 1 : %s", err)
	}

	n, err = api.Crud.Home.Count(ctx, home.QueryFilter{})
	if err != nil {
		t.Fatalf("Should be able to retrieve home count %q : %s", homeType, err)
	}

	if len(hme2) == 0 || len(hme2) != n {
		t.Logf("got: %v", len(hme2))
		t.Logf("exp: %v", n)
		t.Fatalf("Should have 2 homes for page ")
	}

	if hme2[0].ID == hme2[1].ID {
		t.Logf("home1: %v", hme2[0].ID)
		t.Logf("home2: %v", hme2[1].ID)
		t.Fatalf("Should have different home")
	}
}
