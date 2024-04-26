package tests

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/app/api/auth"
	"github.com/ardanlabs/service/business/api/dbtest"
	"github.com/ardanlabs/service/business/domain/homebus"
	"github.com/ardanlabs/service/business/domain/userbus"
)

func insertHomeSeed(dbTest *dbtest.Test, auth *auth.Auth) (appSeedData, error) {
	ctx := context.Background()
	busDomain := dbTest.BusDomain

	usrs, err := userbus.TestGenerateSeedUsers(ctx, 1, userbus.RoleUser, busDomain.User)
	if err != nil {
		return appSeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	hmes, err := homebus.TestGenerateSeedHomes(ctx, 2, busDomain.Home, usrs[0].ID)
	if err != nil {
		return appSeedData{}, fmt.Errorf("seeding homes : %w", err)
	}

	tu1 := appTestUser{
		User: dbtest.User{
			User:  usrs[0],
			Homes: hmes,
		},
		Token: token(dbTest, auth, usrs[0].Email.Address),
	}

	// -------------------------------------------------------------------------

	usrs, err = userbus.TestGenerateSeedUsers(ctx, 1, userbus.RoleUser, busDomain.User)
	if err != nil {
		return appSeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu2 := appTestUser{
		User: dbtest.User{
			User: usrs[0],
		},
		Token: token(dbTest, auth, usrs[0].Email.Address),
	}

	// -------------------------------------------------------------------------

	usrs, err = userbus.TestGenerateSeedUsers(ctx, 1, userbus.RoleAdmin, busDomain.User)
	if err != nil {
		return appSeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	hmes, err = homebus.TestGenerateSeedHomes(ctx, 2, busDomain.Home, usrs[0].ID)
	if err != nil {
		return appSeedData{}, fmt.Errorf("seeding homes : %w", err)
	}

	tu3 := appTestUser{
		User: dbtest.User{
			User:  usrs[0],
			Homes: hmes,
		},
		Token: token(dbTest, auth, usrs[0].Email.Address),
	}

	// -------------------------------------------------------------------------

	usrs, err = userbus.TestGenerateSeedUsers(ctx, 1, userbus.RoleAdmin, busDomain.User)
	if err != nil {
		return appSeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu4 := appTestUser{
		User: dbtest.User{
			User: usrs[0],
		},
		Token: token(dbTest, auth, usrs[0].Email.Address),
	}

	// -------------------------------------------------------------------------

	sd := appSeedData{
		Users:  []appTestUser{tu1, tu2},
		Admins: []appTestUser{tu3, tu4},
	}

	return sd, nil
}
