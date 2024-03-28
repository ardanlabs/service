package tests

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/data/dbtest"
)

func insertUserSeed(dbTest *dbtest.Test) (seedData, error) {
	ctx := context.Background()
	api := dbTest.Core.Crud

	usrs, err := user.TestGenerateSeedUsers(ctx, 2, user.RoleAdmin, api.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu1 := testUser{
		User:  usrs[0],
		token: dbTest.Token(usrs[0].Email.Address),
	}

	tu2 := testUser{
		User:  usrs[1],
		token: dbTest.Token(usrs[1].Email.Address),
	}

	// -------------------------------------------------------------------------

	usrs, err = user.TestGenerateSeedUsers(ctx, 2, user.RoleUser, api.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu3 := testUser{
		User:  usrs[0],
		token: dbTest.Token(usrs[0].Email.Address),
	}

	tu4 := testUser{
		User:  usrs[1],
		token: dbTest.Token(usrs[1].Email.Address),
	}

	// -------------------------------------------------------------------------

	sd := seedData{
		users:  []testUser{tu3, tu4},
		admins: []testUser{tu1, tu2},
	}

	return sd, nil
}
