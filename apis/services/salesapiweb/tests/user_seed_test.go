package tests

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/data/dbtest"
)

func insertUserSeed(dbTest *dbtest.Test) (dbtest.SeedData, error) {
	ctx := context.Background()
	api := dbTest.Core.BusCrud

	usrs, err := user.TestGenerateSeedUsers(ctx, 2, user.RoleAdmin, api.User)
	if err != nil {
		return dbtest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu1 := dbtest.User{
		User:  usrs[0],
		Token: dbTest.Token(usrs[0].Email.Address),
	}

	tu2 := dbtest.User{
		User:  usrs[1],
		Token: dbTest.Token(usrs[1].Email.Address),
	}

	// -------------------------------------------------------------------------

	usrs, err = user.TestGenerateSeedUsers(ctx, 2, user.RoleUser, api.User)
	if err != nil {
		return dbtest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu3 := dbtest.User{
		User:  usrs[0],
		Token: dbTest.Token(usrs[0].Email.Address),
	}

	tu4 := dbtest.User{
		User:  usrs[1],
		Token: dbTest.Token(usrs[1].Email.Address),
	}

	// -------------------------------------------------------------------------

	sd := dbtest.SeedData{
		Users:  []dbtest.User{tu3, tu4},
		Admins: []dbtest.User{tu1, tu2},
	}

	return sd, nil
}
