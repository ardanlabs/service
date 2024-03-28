package tests

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/data/dbtest"
)

func insertVProductSeed(dbTest *dbtest.Test) (seedData, error) {
	ctx := context.Background()
	api := dbTest.Core.Crud

	usrs, err := user.TestGenerateSeedUsers(ctx, 1, user.RoleUser, api.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	prds, err := product.TestGenerateSeedProducts(ctx, 2, api.Product, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding products : %w", err)
	}

	tu1 := testUser{
		User:     usrs[0],
		token:    dbTest.Token(usrs[0].Email.Address, fmt.Sprintf("Password%s", usrs[0].Name[4:])),
		products: prds,
	}

	// -------------------------------------------------------------------------

	usrs, err = user.TestGenerateSeedUsers(ctx, 1, user.RoleAdmin, api.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	prds, err = product.TestGenerateSeedProducts(ctx, 2, api.Product, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding products : %w", err)
	}

	tu2 := testUser{
		User:     usrs[0],
		token:    dbTest.Token(usrs[0].Email.Address, fmt.Sprintf("Password%s", usrs[0].Name[4:])),
		products: prds,
	}

	// -------------------------------------------------------------------------

	sd := seedData{
		admins: []testUser{tu2},
		users:  []testUser{tu1},
	}

	return sd, nil
}
