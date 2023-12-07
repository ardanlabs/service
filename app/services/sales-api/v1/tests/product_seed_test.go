package tests

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/web/v1/order"
)

func createProductSeed(ctx context.Context, dbTest *dbtest.Test) (seedData, error) {
	usrs, err := dbTest.CoreAPIs.User.Query(ctx, user.QueryFilter{}, order.By{Field: user.OrderByName, Direction: order.ASC}, 1, 2)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	// -------------------------------------------------------------------------

	tu1 := testUser{
		User:  usrs[0],
		token: dbTest.TokenV1(usrs[0].Email.Address, "gophers"),
	}

	tu2 := testUser{
		User:  usrs[1],
		token: dbTest.TokenV1(usrs[1].Email.Address, "gophers"),
	}

	// -------------------------------------------------------------------------

	usrs, err = user.TestGenerateSeedUsers(1, user.RoleUser, dbTest.CoreAPIs.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	prds, err := product.TestGenerateSeedProducts(2, dbTest.CoreAPIs.Product, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding products : %w", err)
	}

	tu3 := testUser{
		User:     usrs[0],
		token:    dbTest.TokenV1(usrs[0].Email.Address, fmt.Sprintf("Password%s", usrs[0].Name[4:])),
		products: prds,
	}

	// -------------------------------------------------------------------------

	usrs, err = user.TestGenerateSeedUsers(1, user.RoleAdmin, dbTest.CoreAPIs.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	prds, err = product.TestGenerateSeedProducts(2, dbTest.CoreAPIs.Product, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding products : %w", err)
	}

	tu4 := testUser{
		User:     usrs[0],
		token:    dbTest.TokenV1(usrs[0].Email.Address, fmt.Sprintf("Password%s", usrs[0].Name[4:])),
		products: prds,
	}

	// -------------------------------------------------------------------------

	sd := seedData{
		admins: []testUser{tu1, tu4},
		users:  []testUser{tu2, tu3},
	}

	return sd, nil
}
