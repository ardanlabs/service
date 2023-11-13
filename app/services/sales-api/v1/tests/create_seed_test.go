package tests

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/data/order"
)

func createSeed(ctx context.Context, dbTest *dbtest.Test) (seedData, error) {
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

	sd := seedData{
		admins: []testUser{tu1},
		users:  []testUser{tu2},
	}

	return sd, nil
}
