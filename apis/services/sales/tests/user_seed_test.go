package tests

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/app/api/auth"
	"github.com/ardanlabs/service/business/api/dbtest"
	"github.com/ardanlabs/service/business/domain/userbus"
)

func insertUserSeed(dbTest *dbtest.Test, auth *auth.Auth) (appSeedData, error) {
	ctx := context.Background()
	busDomain := dbTest.BusDomain

	usrs, err := userbus.TestGenerateSeedUsers(ctx, 2, userbus.RoleAdmin, busDomain.User)
	if err != nil {
		return appSeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu1 := appTestUser{
		User: dbtest.User{
			User: usrs[0],
		},
		Token: token(dbTest, auth, usrs[0].Email.Address),
	}

	tu2 := appTestUser{
		User: dbtest.User{
			User: usrs[1],
		},
		Token: token(dbTest, auth, usrs[1].Email.Address),
	}

	// -------------------------------------------------------------------------

	usrs, err = userbus.TestGenerateSeedUsers(ctx, 3, userbus.RoleUser, busDomain.User)
	if err != nil {
		return appSeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu3 := appTestUser{
		User: dbtest.User{
			User: usrs[0],
		},
		Token: token(dbTest, auth, usrs[0].Email.Address),
	}

	tu4 := appTestUser{
		User: dbtest.User{
			User: usrs[1],
		},
		Token: token(dbTest, auth, usrs[1].Email.Address),
	}

	tu5 := appTestUser{
		User: dbtest.User{
			User: usrs[2],
		},
		Token: token(dbTest, auth, usrs[2].Email.Address),
	}

	// -------------------------------------------------------------------------

	sd := appSeedData{
		Users:  []appTestUser{tu3, tu4, tu5},
		Admins: []appTestUser{tu1, tu2},
	}

	return sd, nil
}
