package tests

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/api/http/api/apitest"
	"github.com/ardanlabs/service/app/api/auth"
	"github.com/ardanlabs/service/business/api/dbtest"
	"github.com/ardanlabs/service/business/domain/homebus"
	"github.com/ardanlabs/service/business/domain/userbus"
)

func insertHomeSeed(dbTest *dbtest.Test, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := dbTest.BusDomain

	usrs, err := userbus.TestGenerateSeedUsers(ctx, 1, userbus.RoleUser, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	hmes, err := homebus.TestGenerateSeedHomes(ctx, 2, busDomain.Home, usrs[0].ID)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding homes : %w", err)
	}

	tu1 := apitest.User{
		User: dbtest.User{
			User:  usrs[0],
			Homes: hmes,
		},
		Token: apitest.Token(dbTest, ath, usrs[0].Email.Address),
	}

	// -------------------------------------------------------------------------

	usrs, err = userbus.TestGenerateSeedUsers(ctx, 1, userbus.RoleUser, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu2 := apitest.User{
		User: dbtest.User{
			User: usrs[0],
		},
		Token: apitest.Token(dbTest, ath, usrs[0].Email.Address),
	}

	// -------------------------------------------------------------------------

	usrs, err = userbus.TestGenerateSeedUsers(ctx, 1, userbus.RoleAdmin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	hmes, err = homebus.TestGenerateSeedHomes(ctx, 2, busDomain.Home, usrs[0].ID)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding homes : %w", err)
	}

	tu3 := apitest.User{
		User: dbtest.User{
			User:  usrs[0],
			Homes: hmes,
		},
		Token: apitest.Token(dbTest, ath, usrs[0].Email.Address),
	}

	// -------------------------------------------------------------------------

	usrs, err = userbus.TestGenerateSeedUsers(ctx, 1, userbus.RoleAdmin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu4 := apitest.User{
		User: dbtest.User{
			User: usrs[0],
		},
		Token: apitest.Token(dbTest, ath, usrs[0].Email.Address),
	}

	// -------------------------------------------------------------------------

	sd := apitest.SeedData{
		Users:  []apitest.User{tu1, tu2},
		Admins: []apitest.User{tu3, tu4},
	}

	return sd, nil
}
