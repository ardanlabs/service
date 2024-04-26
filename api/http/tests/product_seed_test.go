package tests

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/app/api/auth"
	"github.com/ardanlabs/service/business/api/dbtest"
	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/ardanlabs/service/business/domain/userbus"
)

func insertProductSeed(dbTest *dbtest.Test, auth *auth.Auth) (appSeedData, error) {
	ctx := context.Background()
	busDomain := dbTest.BusDomain

	usrs, err := userbus.TestGenerateSeedUsers(ctx, 1, userbus.RoleUser, busDomain.User)
	if err != nil {
		return appSeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	prds, err := productbus.TestGenerateSeedProducts(ctx, 2, busDomain.Product, usrs[0].ID)
	if err != nil {
		return appSeedData{}, fmt.Errorf("seeding products : %w", err)
	}

	tu1 := appTestUser{
		User: dbtest.User{
			User:     usrs[0],
			Products: prds,
		},
		Token: token(dbTest, auth, usrs[0].Email.Address),
	}

	// -------------------------------------------------------------------------

	usrs, err = userbus.TestGenerateSeedUsers(ctx, 1, userbus.RoleAdmin, busDomain.User)
	if err != nil {
		return appSeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	prds, err = productbus.TestGenerateSeedProducts(ctx, 2, busDomain.Product, usrs[0].ID)
	if err != nil {
		return appSeedData{}, fmt.Errorf("seeding products : %w", err)
	}

	tu2 := appTestUser{
		User: dbtest.User{
			User:     usrs[0],
			Products: prds,
		},
		Token: token(dbTest, auth, usrs[0].Email.Address),
	}

	// -------------------------------------------------------------------------

	sd := appSeedData{
		Admins: []appTestUser{tu2},
		Users:  []appTestUser{tu1},
	}

	return sd, nil
}
