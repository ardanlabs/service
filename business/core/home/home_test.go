package home_test

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/core/home"
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

func Test_Home(t *testing.T) {
	t.Run("crud", crud)
	t.Run("paging", paging)
}

func crud(t *testing.T) {
	seed := func(ctx context.Context, usrCore *user.Core, hmeCore *home.Core) ([]home.Home, error) {
		var filter user.QueryFilter
		filter.WithName("Admin Gopher")

		usrs, err := usrCore.Query(ctx, filter, user.DefaultOrderBy, 1, 1)
		if err != nil {
			return nil, fmt.Errorf("seeding users : %w", err)
		}

		hmes, err := home.TestGenerateSeedHomes(1, hmeCore, usrs[0].ID)
		if err != nil {
			return nil, fmt.Errorf("seeding homes : %w", err)
		}

		return hmes, nil
	}

	// ---------------------------------------------------------------------------

	test := dbtest.NewTest(t, c, "Test_Home/crud")

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

	hmes, err := seed(ctx, api.User, api.Home)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// ---------------------------------------------------------------------------

	saved, err := api.Home.QueryByID(ctx, hmes[0].ID)
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

	if _, err := api.Home.Update(ctx, saved, upd); err != nil {
		t.Errorf("Should be able to update home : %s", err)
	}

	saved, err = api.Home.QueryByID(ctx, hmes[0].ID)
	if err != nil {
		t.Fatalf("Should be able to retrieve updated home : %s", err)
	}

	diff := hmes[0].DateUpdated.Sub(saved.DateUpdated)
	if diff > 0 {
		t.Fatalf("Should have a larger DateUpdated : sav %v, hme %v, dif %v", saved.DateUpdated, saved.DateUpdated, diff)
	}

	homes, err := api.Home.Query(ctx, home.QueryFilter{}, user.DefaultOrderBy, 1, 3)
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

	if _, err := api.Home.Update(ctx, saved, upd); err != nil {
		t.Fatalf("Should be able to update just some fields of home : %s", err)
	}

	saved, err = api.Home.QueryByID(ctx, hmes[0].ID)
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

	if err := api.Home.Delete(ctx, saved); err != nil {
		t.Fatalf("Should be able to delete home : %s", err)
	}

	_, err = api.Home.QueryByID(ctx, hmes[0].ID)
	if !errors.Is(err, home.ErrNotFound) {
		t.Fatalf("Should NOT be able to retrieve deleted home : %s", err)
	}
}

func paging(t *testing.T) {
	seed := func(ctx context.Context, usrCore *user.Core, hmeCore *home.Core) ([]home.Home, error) {
		var filter user.QueryFilter
		filter.WithName("Admin Gopher")

		usrs, err := usrCore.Query(ctx, filter, user.DefaultOrderBy, 1, 1)
		if err != nil {
			return nil, fmt.Errorf("seeding homes : %w", err)
		}

		hmes, err := home.TestGenerateSeedHomes(2, hmeCore, usrs[0].ID)
		if err != nil {
			return nil, fmt.Errorf("seeding homes : %w", err)
		}

		return hmes, nil
	}

	// -------------------------------------------------------------------------

	test := dbtest.NewTest(t, c, "Test_Home/paging")
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

	hmes, err := seed(ctx, api.User, api.Home)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	homeType := hmes[0].Type
	hme1, err := api.Home.Query(ctx, home.QueryFilter{Type: &homeType}, user.DefaultOrderBy, 1, 2)
	if err != nil {
		t.Fatalf("Should be able to retrieve homes %q : %s", homeType, err)
	}

	n, err := api.Home.Count(ctx, home.QueryFilter{Type: &homeType})
	if err != nil {
		t.Fatalf("Should be able to retrieve home count %q : %s", homeType, err)
	}

	if len(hme1) != n {
		t.Log("got:", len(hme1))
		t.Log("exp:", n)
		t.Fatal("Should have the correct number of homes")
	}

	hme2, err := api.Home.Query(ctx, home.QueryFilter{}, user.DefaultOrderBy, 1, 2)
	if err != nil {
		t.Fatalf("Should be able to retrieve 2 homes for page 1 : %s", err)
	}

	n, err = api.Home.Count(ctx, home.QueryFilter{})
	if err != nil {
		t.Fatalf("Should be able to retrieve home count %q : %s", homeType, err)
	}

	if len(hme2) != n {
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
