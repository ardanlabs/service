package tests

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/domain/userbus"
)

func insertUserSeed(dbTest *dbtest.Test) (dbtest.SeedData, error) {
	ctx := context.Background()
	api := dbTest.Core.BusCrud

	usrs, err := userbus.TestGenerateSeedUsers(ctx, 2, userbus.RoleAdmin, api.User)
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

	usrs, err = userbus.TestGenerateSeedUsers(ctx, 2, userbus.RoleUser, api.User)
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
