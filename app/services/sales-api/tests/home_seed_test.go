package tests

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/business/core/crud/home"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/data/dbtest"
)

func insertHomeSeed(dbTest *dbtest.Test) (dbtest.SeedData, error) {
	ctx := context.Background()
	api := dbTest.Core.Crud

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
