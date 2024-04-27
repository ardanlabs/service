package tests

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/app/api/auth"
	"github.com/ardanlabs/service/business/api/dbtest"
	"github.com/ardanlabs/service/business/domain/homebus"
	"github.com/ardanlabs/service/business/domain/userbus"
)

func insertHomeSeed(dbTest *dbtest.Test, auth *auth.Auth) (seedData, error) {
	ctx := context.Background()
	busDomain := dbTest.BusDomain

	usrs, err := userbus.TestGenerateSeedUsers(ctx, 1, userbus.RoleUser, busDomain.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	hmes, err := homebus.TestGenerateSeedHomes(ctx, 2, busDomain.Home, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding homes : %w", err)
	}

	tu1 := testUser{
		User: dbtest.User{
			User:  usrs[0],
			Homes: hmes,
		},
		Token: token(dbTest, auth, usrs[0].Email.Address),
	}

	// -------------------------------------------------------------------------

	usrs, err = userbus.TestGenerateSeedUsers(ctx, 1, userbus.RoleUser, busDomain.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu2 := testUser{
		User: dbtest.User{
			User: usrs[0],
		},
		Token: token(dbTest, auth, usrs[0].Email.Address),
	}

	// -------------------------------------------------------------------------

	usrs, err = userbus.TestGenerateSeedUsers(ctx, 1, userbus.RoleAdmin, busDomain.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	hmes, err = homebus.TestGenerateSeedHomes(ctx, 2, busDomain.Home, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding homes : %w", err)
	}

	tu3 := testUser{
		User: dbtest.User{
			User:  usrs[0],
			Homes: hmes,
		},
		Token: token(dbTest, auth, usrs[0].Email.Address),
	}

	// -------------------------------------------------------------------------

	usrs, err = userbus.TestGenerateSeedUsers(ctx, 1, userbus.RoleAdmin, busDomain.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu4 := testUser{
		User: dbtest.User{
			User: usrs[0],
		},
		Token: token(dbTest, auth, usrs[0].Email.Address),
	}

	// -------------------------------------------------------------------------

	sd := seedData{
		Users:  []testUser{tu1, tu2},
		Admins: []testUser{tu3, tu4},
	}

	return sd, nil
}
