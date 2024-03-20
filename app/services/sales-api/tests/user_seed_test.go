package tests

import (
	"fmt"

	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/data/dbtest"
)

func createUserSeed(dbTest *dbtest.Test) (seedData, error) {
	usrs, err := user.TestGenerateSeedUsers(2, user.RoleAdmin, dbTest.CoreAPIs.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu1 := testUser{
		User:  usrs[0],
		token: dbTest.TokenV1(usrs[0].Email.Address, fmt.Sprintf("Password%s", usrs[0].Name[4:])),
	}

	tu2 := testUser{
		User:  usrs[1],
		token: dbTest.TokenV1(usrs[1].Email.Address, fmt.Sprintf("Password%s", usrs[1].Name[4:])),
	}

	// -------------------------------------------------------------------------

	usrs, err = user.TestGenerateSeedUsers(2, user.RoleUser, dbTest.CoreAPIs.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu3 := testUser{
		User:  usrs[0],
		token: dbTest.TokenV1(usrs[0].Email.Address, fmt.Sprintf("Password%s", usrs[0].Name[4:])),
	}

	tu4 := testUser{
		User:  usrs[1],
		token: dbTest.TokenV1(usrs[1].Email.Address, fmt.Sprintf("Password%s", usrs[1].Name[4:])),
	}

	// -------------------------------------------------------------------------

	sd := seedData{
		users:  []testUser{tu3, tu4},
		admins: []testUser{tu1, tu2},
	}

	return sd, nil
}
