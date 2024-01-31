package tests

import (
	"fmt"

	"github.com/ardanlabs/service/business/core/crud/home"
	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/data/dbtest"
)

func createHomeSeed(dbTest *dbtest.Test) (seedData, error) {
	usrs, err := user.TestGenerateSeedUsers(1, user.RoleUser, dbTest.CoreAPIs.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	prds, err := product.TestGenerateSeedProducts(2, dbTest.CoreAPIs.Product, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding products : %w", err)
	}

	hmes, err := home.TestGenerateSeedHomes(2, dbTest.CoreAPIs.Home, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding homes : %w", err)
	}

	tu1 := testUser{
		User:     usrs[0],
		token:    dbTest.TokenV1(usrs[0].Email.Address, fmt.Sprintf("Password%s", usrs[0].Name[4:])),
		products: prds,
		homes:    hmes,
	}

	// -------------------------------------------------------------------------

	usrs, err = user.TestGenerateSeedUsers(1, user.RoleUser, dbTest.CoreAPIs.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu2 := testUser{
		User:  usrs[0],
		token: dbTest.TokenV1(usrs[0].Email.Address, fmt.Sprintf("Password%s", usrs[0].Name[4:])),
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

	hmes, err = home.TestGenerateSeedHomes(2, dbTest.CoreAPIs.Home, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding homes : %w", err)
	}

	tu3 := testUser{
		User:     usrs[0],
		token:    dbTest.TokenV1(usrs[0].Email.Address, fmt.Sprintf("Password%s", usrs[0].Name[4:])),
		products: prds,
		homes:    hmes,
	}

	// -------------------------------------------------------------------------

	usrs, err = user.TestGenerateSeedUsers(1, user.RoleAdmin, dbTest.CoreAPIs.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu4 := testUser{
		User:  usrs[0],
		token: dbTest.TokenV1(usrs[0].Email.Address, fmt.Sprintf("Password%s", usrs[0].Name[4:])),
	}

	// -------------------------------------------------------------------------

	sd := seedData{
		users:  []testUser{tu1, tu2},
		admins: []testUser{tu3, tu4},
	}

	return sd, nil
}
