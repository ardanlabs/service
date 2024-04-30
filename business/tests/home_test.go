package tests

import (
	"context"
	"fmt"
	"runtime/debug"
	"sort"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/api/dbtest"
	"github.com/ardanlabs/service/business/api/unittest"
	"github.com/ardanlabs/service/business/domain/homebus"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/google/go-cmp/cmp"
)

func Test_Home(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, c, "Test_Home")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		db.Teardown()
	}()

	sd, err := insertHomeSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	unittest.Run(t, homeQuery(db.BusDomain, sd), "home-query")
	unittest.Run(t, homeCreate(db.BusDomain, sd), "home-create")
	unittest.Run(t, homeUpdate(db.BusDomain, sd), "home-update")
	unittest.Run(t, homeDelete(db.BusDomain, sd), "home-delete")
}

// =============================================================================

func insertHomeSeedData(busDomain dbtest.BusDomain) (unittest.SeedData, error) {
	ctx := context.Background()

	usrs, err := userbus.TestSeedUsers(ctx, 1, userbus.RoleUser, busDomain.User)
	if err != nil {
		return unittest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	hmes, err := homebus.TestGenerateSeedHomes(ctx, 2, busDomain.Home, usrs[0].ID)
	if err != nil {
		return unittest.SeedData{}, fmt.Errorf("seeding homes : %w", err)
	}

	tu1 := unittest.User{
		User:  usrs[0],
		Homes: hmes,
	}

	// -------------------------------------------------------------------------

	usrs, err = userbus.TestSeedUsers(ctx, 1, userbus.RoleUser, busDomain.User)
	if err != nil {
		return unittest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu2 := unittest.User{
		User: usrs[0],
	}

	// -------------------------------------------------------------------------

	usrs, err = userbus.TestSeedUsers(ctx, 1, userbus.RoleAdmin, busDomain.User)
	if err != nil {
		return unittest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	hmes, err = homebus.TestGenerateSeedHomes(ctx, 2, busDomain.Home, usrs[0].ID)
	if err != nil {
		return unittest.SeedData{}, fmt.Errorf("seeding homes : %w", err)
	}

	tu3 := unittest.User{
		User:  usrs[0],
		Homes: hmes,
	}

	// -------------------------------------------------------------------------

	usrs, err = userbus.TestSeedUsers(ctx, 1, userbus.RoleAdmin, busDomain.User)
	if err != nil {
		return unittest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu4 := unittest.User{
		User: usrs[0],
	}

	// -------------------------------------------------------------------------

	sd := unittest.SeedData{
		Users:  []unittest.User{tu1, tu2},
		Admins: []unittest.User{tu3, tu4},
	}

	return sd, nil
}

// =============================================================================

func homeQuery(busDomain dbtest.BusDomain, sd unittest.SeedData) []unittest.Table {
	hmes := make([]homebus.Home, 0, len(sd.Admins[0].Homes)+len(sd.Users[0].Homes))
	hmes = append(hmes, sd.Admins[0].Homes...)
	hmes = append(hmes, sd.Users[0].Homes...)

	sort.Slice(hmes, func(i, j int) bool {
		return hmes[i].ID.String() <= hmes[j].ID.String()
	})

	table := []unittest.Table{
		{
			Name:    "all",
			ExpResp: hmes,
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Home.Query(ctx, homebus.QueryFilter{}, homebus.DefaultOrderBy, 1, 10)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.([]homebus.Home)
				if !exists {
					return "error occurred"
				}

				expResp := exp.([]homebus.Home)

				for i := range gotResp {
					if gotResp[i].DateCreated.Format(time.RFC3339) == expResp[i].DateCreated.Format(time.RFC3339) {
						expResp[i].DateCreated = gotResp[i].DateCreated
					}

					if gotResp[i].DateUpdated.Format(time.RFC3339) == expResp[i].DateUpdated.Format(time.RFC3339) {
						expResp[i].DateUpdated = gotResp[i].DateUpdated
					}
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:    "byid",
			ExpResp: sd.Users[0].Homes[0],
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Home.QueryByID(ctx, sd.Users[0].Homes[0].ID)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(homebus.Home)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(homebus.Home)

				if gotResp.DateCreated.Format(time.RFC3339) == expResp.DateCreated.Format(time.RFC3339) {
					expResp.DateCreated = gotResp.DateCreated
				}

				if gotResp.DateUpdated.Format(time.RFC3339) == expResp.DateUpdated.Format(time.RFC3339) {
					expResp.DateUpdated = gotResp.DateUpdated
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func homeCreate(busDomain dbtest.BusDomain, sd unittest.SeedData) []unittest.Table {
	table := []unittest.Table{
		{
			Name: "basic",
			ExpResp: homebus.Home{
				UserID: sd.Users[0].ID,
				Type:   homebus.TypeSingle,
				Address: homebus.Address{
					Address1: "123 Mocking Bird Lane",
					ZipCode:  "35810",
					City:     "Huntsville",
					State:    "AL",
					Country:  "US",
				},
			},
			ExcFunc: func(ctx context.Context) any {
				nh := homebus.NewHome{
					UserID: sd.Users[0].ID,
					Type:   homebus.TypeSingle,
					Address: homebus.Address{
						Address1: "123 Mocking Bird Lane",
						ZipCode:  "35810",
						City:     "Huntsville",
						State:    "AL",
						Country:  "US",
					},
				}

				resp, err := busDomain.Home.Create(ctx, nh)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(homebus.Home)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(homebus.Home)

				expResp.ID = gotResp.ID
				expResp.DateCreated = gotResp.DateCreated
				expResp.DateUpdated = gotResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func homeUpdate(busDomain dbtest.BusDomain, sd unittest.SeedData) []unittest.Table {
	table := []unittest.Table{
		{
			Name: "basic",
			ExpResp: homebus.Home{
				ID:     sd.Users[0].Homes[0].ID,
				UserID: sd.Users[0].ID,
				Type:   homebus.TypeSingle,
				Address: homebus.Address{
					Address1: "123 Mocking Bird Lane",
					Address2: "apt 105",
					ZipCode:  "35810",
					City:     "Huntsville",
					State:    "AL",
					Country:  "US",
				},
				DateCreated: sd.Users[0].Homes[0].DateCreated,
				DateUpdated: sd.Users[0].Homes[0].DateCreated,
			},
			ExcFunc: func(ctx context.Context) any {
				uh := homebus.UpdateHome{
					Type: &homebus.TypeSingle,
					Address: &homebus.UpdateAddress{
						Address1: dbtest.StringPointer("123 Mocking Bird Lane"),
						Address2: dbtest.StringPointer("apt 105"),
						ZipCode:  dbtest.StringPointer("35810"),
						City:     dbtest.StringPointer("Huntsville"),
						State:    dbtest.StringPointer("AL"),
						Country:  dbtest.StringPointer("US"),
					},
				}

				resp, err := busDomain.Home.Update(ctx, sd.Users[0].Homes[0], uh)
				if err != nil {
					return err
				}

				resp.DateUpdated = resp.DateCreated

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(homebus.Home)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(homebus.Home)

				expResp.DateUpdated = gotResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func homeDelete(busDomain dbtest.BusDomain, sd unittest.SeedData) []unittest.Table {
	table := []unittest.Table{
		{
			Name:    "user",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.Home.Delete(ctx, sd.Users[0].Homes[1]); err != nil {
					return err
				}

				return nil
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "admin",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.Home.Delete(ctx, sd.Admins[0].Homes[1]); err != nil {
					return err
				}

				return nil
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
