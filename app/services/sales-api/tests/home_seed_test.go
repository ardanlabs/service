package tests

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/business/core/crud/home"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/data/dbtest"
)

func insertHomeSeed(dbTest *dbtest.Test) (seedData, error) {
	ctx := context.Background()
	api := dbTest.Core.Crud

	usrs, err := user.TestGenerateSeedUsers(ctx, 1, user.RoleUser, api.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	hmes, err := home.TestGenerateSeedHomes(ctx, 2, api.Home, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding homes : %w", err)
	}

	tu1 := testUser{
		User:  usrs[0],
		token: dbTest.TokenV1(usrs[0].Email.Address, fmt.Sprintf("Password%s", usrs[0].Name[4:])),
		homes: hmes,
	}

	// -------------------------------------------------------------------------

	usrs, err = user.TestGenerateSeedUsers(ctx, 1, user.RoleUser, api.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu2 := testUser{
		User:  usrs[0],
		token: dbTest.TokenV1(usrs[0].Email.Address, fmt.Sprintf("Password%s", usrs[0].Name[4:])),
	}

	// -------------------------------------------------------------------------

	usrs, err = user.TestGenerateSeedUsers(ctx, 1, user.RoleAdmin, api.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	hmes, err = home.TestGenerateSeedHomes(ctx, 2, api.Home, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding homes : %w", err)
	}

	tu3 := testUser{
		User:  usrs[0],
		token: dbTest.TokenV1(usrs[0].Email.Address, fmt.Sprintf("Password%s", usrs[0].Name[4:])),
		homes: hmes,
	}

	// -------------------------------------------------------------------------

	usrs, err = user.TestGenerateSeedUsers(ctx, 1, user.RoleAdmin, api.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu4 := testUser{
		User:  usrs[0],
		token: dbTest.TokenV1(usrs[0].Email.Address, fmt.Sprintf("Password%s", usrs[0].Name[4:])),
	}

	// -------------------------------------------------------------------------

	sd := seedData{
		users:  []testUser{tu1, tu2},
		admins: []testUser{tu3, tu4},
	}

	return sd, nil
}
