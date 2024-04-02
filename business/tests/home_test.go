package tests

import (
	"context"
	"fmt"
	"runtime/debug"
	"sort"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/core/crud/home"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/google/go-cmp/cmp"
)

func Test_Home(t *testing.T) {
	t.Parallel()

	dbTest := dbtest.NewTest(t, c, "Test_Home")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		dbTest.Teardown()
	}()

	sd, err := insertHomeSeedData(dbTest)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	dbtest.UnitTest(t, homeQuery(dbTest, sd), "home-query")
	dbtest.UnitTest(t, homeCreate(dbTest, sd), "home-create")
	dbtest.UnitTest(t, homeUpdate(dbTest, sd), "home-update")
	dbtest.UnitTest(t, homeDelete(dbTest, sd), "home-delete")
}

// =============================================================================

func insertHomeSeedData(dbTest *dbtest.Test) (dbtest.SeedData, error) {
	ctx := context.Background()
	api := dbTest.Core.BusCrud

	usrs, err := user.TestGenerateSeedUsers(ctx, 1, user.RoleUser, api.User)
	if err != nil {
		return dbtest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	hmes, err := home.TestGenerateSeedHomes(ctx, 2, api.Home, usrs[0].ID)
	if err != nil {
		return dbtest.SeedData{}, fmt.Errorf("seeding homes : %w", err)
	}

	tu1 := dbtest.User{
		User:  usrs[0],
		Token: dbTest.Token(usrs[0].Email.Address),
		Homes: hmes,
	}

	// -------------------------------------------------------------------------

	usrs, err = user.TestGenerateSeedUsers(ctx, 1, user.RoleUser, api.User)
	if err != nil {
		return dbtest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu2 := dbtest.User{
		User:  usrs[0],
		Token: dbTest.Token(usrs[0].Email.Address),
	}

	// -------------------------------------------------------------------------

	usrs, err = user.TestGenerateSeedUsers(ctx, 1, user.RoleAdmin, api.User)
	if err != nil {
		return dbtest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	hmes, err = home.TestGenerateSeedHomes(ctx, 2, api.Home, usrs[0].ID)
	if err != nil {
		return dbtest.SeedData{}, fmt.Errorf("seeding homes : %w", err)
	}

	tu3 := dbtest.User{
		User:  usrs[0],
		Token: dbTest.Token(usrs[0].Email.Address),
		Homes: hmes,
	}

	// -------------------------------------------------------------------------

	usrs, err = user.TestGenerateSeedUsers(ctx, 1, user.RoleAdmin, api.User)
	if err != nil {
		return dbtest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu4 := dbtest.User{
		User:  usrs[0],
		Token: dbTest.Token(usrs[0].Email.Address),
	}

	// -------------------------------------------------------------------------

	sd := dbtest.SeedData{
		Users:  []dbtest.User{tu1, tu2},
		Admins: []dbtest.User{tu3, tu4},
	}

	return sd, nil
}

// =============================================================================

func homeQuery(dbt *dbtest.Test, sd dbtest.SeedData) []dbtest.UnitTable {
	hmes := make([]home.Home, 0, len(sd.Admins[0].Homes)+len(sd.Users[0].Homes))
	hmes = append(hmes, sd.Admins[0].Homes...)
	hmes = append(hmes, sd.Users[0].Homes...)

	sort.Slice(hmes, func(i, j int) bool {
		return hmes[i].ID.String() <= hmes[j].ID.String()
	})

	table := []dbtest.UnitTable{
		{
			Name:    "all",
			ExpResp: hmes,
			ExcFunc: func(ctx context.Context) any {
				resp, err := dbt.Core.BusCrud.Home.Query(ctx, home.QueryFilter{}, home.DefaultOrderBy, 1, 10)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.([]home.Home)
				if !exists {
					return "error occurred"
				}

				expResp := exp.([]home.Home)

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
				resp, err := dbt.Core.BusCrud.Home.QueryByID(ctx, sd.Users[0].Homes[0].ID)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(home.Home)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(home.Home)

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

func homeCreate(dbt *dbtest.Test, sd dbtest.SeedData) []dbtest.UnitTable {
	table := []dbtest.UnitTable{
		{
			Name: "basic",
			ExpResp: home.Home{
				UserID: sd.Users[0].ID,
				Type:   home.TypeSingle,
				Address: home.Address{
					Address1: "123 Mocking Bird Lane",
					ZipCode:  "35810",
					City:     "Huntsville",
					State:    "AL",
					Country:  "US",
				},
			},
			ExcFunc: func(ctx context.Context) any {
				nh := home.NewHome{
					UserID: sd.Users[0].ID,
					Type:   home.TypeSingle,
					Address: home.Address{
						Address1: "123 Mocking Bird Lane",
						ZipCode:  "35810",
						City:     "Huntsville",
						State:    "AL",
						Country:  "US",
					},
				}

				resp, err := dbt.Core.BusCrud.Home.Create(ctx, nh)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(home.Home)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(home.Home)

				expResp.ID = gotResp.ID
				expResp.DateCreated = gotResp.DateCreated
				expResp.DateUpdated = gotResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func homeUpdate(dbt *dbtest.Test, sd dbtest.SeedData) []dbtest.UnitTable {
	table := []dbtest.UnitTable{
		{
			Name: "basic",
			ExpResp: home.Home{
				ID:     sd.Users[0].Homes[0].ID,
				UserID: sd.Users[0].ID,
				Type:   home.TypeSingle,
				Address: home.Address{
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
				uh := home.UpdateHome{
					Type: &home.TypeSingle,
					Address: &home.UpdateAddress{
						Address1: dbtest.StringPointer("123 Mocking Bird Lane"),
						Address2: dbtest.StringPointer("apt 105"),
						ZipCode:  dbtest.StringPointer("35810"),
						City:     dbtest.StringPointer("Huntsville"),
						State:    dbtest.StringPointer("AL"),
						Country:  dbtest.StringPointer("US"),
					},
				}

				resp, err := dbt.Core.BusCrud.Home.Update(ctx, sd.Users[0].Homes[0], uh)
				if err != nil {
					return err
				}

				resp.DateUpdated = resp.DateCreated

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(home.Home)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(home.Home)

				expResp.DateUpdated = gotResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func homeDelete(dbt *dbtest.Test, sd dbtest.SeedData) []dbtest.UnitTable {
	table := []dbtest.UnitTable{
		{
			Name:    "user",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := dbt.Core.BusCrud.Home.Delete(ctx, sd.Users[0].Homes[1]); err != nil {
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
				if err := dbt.Core.BusCrud.Home.Delete(ctx, sd.Admins[0].Homes[1]); err != nil {
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
