package tests

import (
	"context"
	"fmt"
	"runtime/debug"
	"sort"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/core/views/vproduct"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/google/go-cmp/cmp"
)

func Test_VProduct(t *testing.T) {
	t.Parallel()

	dbTest := dbtest.NewTest(t, c, "Test_Product")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		dbTest.Teardown()
	}()

	sd, err := insertVProductSeedData(dbTest)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	dbtest.UnitTest(t, vproductQuery(dbTest, sd), "vproduct-query")
}

// =============================================================================

func insertVProductSeedData(dbTest *dbtest.Test) (dbtest.SeedData, error) {
	ctx := context.Background()
	api := dbTest.Core.Crud

	usrs, err := user.TestGenerateSeedUsers(ctx, 1, user.RoleUser, api.User)
	if err != nil {
		return dbtest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	prds, err := product.TestGenerateSeedProducts(ctx, 2, api.Product, usrs[0].ID)
	if err != nil {
		return dbtest.SeedData{}, fmt.Errorf("seeding products : %w", err)
	}

	tu1 := dbtest.User{
		User:     usrs[0],
		Token:    dbTest.Token(usrs[0].Email.Address),
		Products: prds,
	}

	// -------------------------------------------------------------------------

	usrs, err = user.TestGenerateSeedUsers(ctx, 1, user.RoleAdmin, api.User)
	if err != nil {
		return dbtest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	prds, err = product.TestGenerateSeedProducts(ctx, 2, api.Product, usrs[0].ID)
	if err != nil {
		return dbtest.SeedData{}, fmt.Errorf("seeding products : %w", err)
	}

	tu2 := dbtest.User{
		User:     usrs[0],
		Token:    dbTest.Token(usrs[0].Email.Address),
		Products: prds,
	}

	// -------------------------------------------------------------------------

	sd := dbtest.SeedData{
		Admins: []dbtest.User{tu2},
		Users:  []dbtest.User{tu1},
	}

	return sd, nil
}

// =============================================================================

func toVProduct(usr user.User, prd product.Product) vproduct.Product {
	return vproduct.Product{
		ID:          prd.ID,
		UserID:      prd.UserID,
		Name:        prd.Name,
		Cost:        prd.Cost,
		Quantity:    prd.Quantity,
		DateCreated: prd.DateCreated,
		DateUpdated: prd.DateUpdated,
		UserName:    usr.Name,
	}
}

func toVProducts(usr user.User, prds []product.Product) []vproduct.Product {
	items := make([]vproduct.Product, len(prds))
	for i, prd := range prds {
		items[i] = toVProduct(usr, prd)
	}

	return items
}

// =============================================================================

func vproductQuery(dbt *dbtest.Test, sd dbtest.SeedData) []dbtest.UnitTable {
	prds := toVProducts(sd.Admins[0].User, sd.Admins[0].Products)
	prds = append(prds, toVProducts(sd.Users[0].User, sd.Users[0].Products)...)

	sort.Slice(prds, func(i, j int) bool {
		return prds[i].ID.String() <= prds[j].ID.String()
	})

	table := []dbtest.UnitTable{
		{
			Name:    "all",
			ExpResp: prds,
			ExcFunc: func(ctx context.Context) any {
				filter := vproduct.QueryFilter{
					Name: dbtest.StringPointer("Name"),
				}

				resp, err := dbt.Core.View.Product.Query(ctx, filter, vproduct.DefaultOrderBy, 1, 10)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.([]vproduct.Product)
				if !exists {
					return "error occurred"
				}

				expResp := exp.([]vproduct.Product)

				for i := range gotResp {
					if gotResp[i].DateCreated.Format(time.RFC3339) == expResp[i].DateCreated.Format(time.RFC3339) {
						expResp[i].DateCreated = gotResp[i].DateCreated
					}

					if gotResp[i].DateUpdated.Format(time.RFC3339) == expResp[i].DateUpdated.Format(time.RFC3339) {
						expResp[i].DateUpdated = gotResp[i].DateUpdated
					}
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}
