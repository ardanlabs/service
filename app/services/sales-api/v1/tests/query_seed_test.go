package tests

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/business/core/home"
	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/data/order"
)

func querySeed(ctx context.Context, dbTest *dbtest.Test) (seedData, error) {
	usrs, err := dbTest.CoreAPIs.User.Query(ctx, user.QueryFilter{}, order.By{Field: user.OrderByName, Direction: order.ASC}, 1, 2)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	// -------------------------------------------------------------------------

	tu1 := testUser{
		User:  usrs[0],
		token: dbTest.TokenV1(usrs[0].Email.Address, "gophers"),
	}

	tu1.products, err = product.TestGenerateSeedProducts(5, dbTest.CoreAPIs.Product, tu1.ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding products1 : %w", err)
	}

	tu1.homes, err = home.TestGenerateSeedHomes(5, dbTest.CoreAPIs.Home, tu1.ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding homes1 : %w", err)
	}

	// -------------------------------------------------------------------------

	tu2 := testUser{
		User:  usrs[1],
		token: dbTest.TokenV1(usrs[1].Email.Address, "gophers"),
	}

	tu2.products, err = product.TestGenerateSeedProducts(5, dbTest.CoreAPIs.Product, tu2.ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding products2 : %w", err)
	}

	tu2.homes, err = home.TestGenerateSeedHomes(5, dbTest.CoreAPIs.Home, tu2.ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding homes2 : %w", err)
	}

	// -------------------------------------------------------------------------

	sd := seedData{
		admins: []testUser{tu1},
		users:  []testUser{tu2},
	}

	return sd, nil
}
