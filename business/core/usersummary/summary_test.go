package usersummary_test

import (
	"context"
	"fmt"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/core/usersummary"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/web/v1/order"
	"github.com/ardanlabs/service/foundation/docker"
)

var c *docker.Container

func TestMain(m *testing.M) {
	var err error
	c, err = dbtest.StartDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dbtest.StopDB(c)

	m.Run()
}

func Test_Product(t *testing.T) {
	t.Run("paging", paging)
}

func paging(t *testing.T) {
	seed := func(ctx context.Context, usrCore *user.Core, prdCore *product.Core) ([]user.User, []product.Product, error) {
		usrs, err := usrCore.Query(ctx, user.QueryFilter{}, order.By{Field: user.OrderByName, Direction: order.ASC}, 1, 2)
		if err != nil {
			return nil, nil, fmt.Errorf("seeding users : %w", err)
		}

		prds1, err := product.TestGenerateSeedProducts(5, prdCore, usrs[0].ID)
		if err != nil {
			return nil, nil, fmt.Errorf("seeding products : %w", err)
		}

		prds2, err := product.TestGenerateSeedProducts(5, prdCore, usrs[1].ID)
		if err != nil {
			return nil, nil, fmt.Errorf("seeding products : %w", err)
		}

		var prds []product.Product
		prds = append(prds, prds1...)
		prds = append(prds, prds2...)

		return usrs, prds, nil
	}

	// -------------------------------------------------------------------------

	test := dbtest.NewTest(t, c, "Test_Product/paging")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.Teardown()
	}()

	api := test.CoreAPIs

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Log("Go seeding ...")

	usrs, prds, err := seed(ctx, api.User, api.Product)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	prd1, err := api.UserSummary.Query(ctx, usersummary.QueryFilter{}, order.By{Field: usersummary.OrderByUserName, Direction: order.ASC}, 1, 10)
	if err != nil {
		t.Fatalf("Should be able to retrieve user summary : %s", err)
	}

	n, err := api.UserSummary.Count(ctx, usersummary.QueryFilter{})
	if err != nil {
		t.Fatalf("Should be able to retrieve user summary count : %s", err)
	}

	if len(prd1) != n && len(prd1) != 2 {
		t.Log("got:", len(prd1))
		t.Log("exp:", n)
		t.Log("exp:", 2)
		t.Fatal("Should all the user summary records")
	}

	// -------------------------------------------------------------------------

	totalCount1 := 5
	var totalCost1 float64
	for i := 0; i < totalCount1; i++ {
		totalCost1 += prds[i].Cost
	}

	totalCount2 := 5
	var totalCost2 float64
	for i := 5; i < totalCount1+totalCount2; i++ {
		totalCost2 += prds[i].Cost
	}

	cases := []usersummary.Summary{
		{
			UserID:     usrs[0].ID,
			UserName:   usrs[0].Name,
			TotalCount: totalCount1,
			TotalCost:  totalCost1,
		},
		{
			UserID:     usrs[1].ID,
			UserName:   usrs[1].Name,
			TotalCount: totalCount2,
			TotalCost:  totalCost2,
		},
	}

	for i, c := range cases {
		t.Run(c.UserName, func(t *testing.T) {
			if prd1[i].UserID != c.UserID {
				t.Log("got:", prd1[i].UserID)
				t.Log("exp:", c.UserID)
				t.Error("Should match the UserID")
			}

			if prd1[i].UserName != c.UserName {
				t.Log("got:", prd1[i].UserName)
				t.Log("exp:", c.UserName)
				t.Error("Should match the UserName")
			}

			if prd1[i].TotalCount != c.TotalCount {
				t.Log("got:", prd1[i].TotalCount)
				t.Log("exp:", c.TotalCount)
				t.Error("Should match the TotalCount")
			}

			if prd1[i].TotalCost != c.TotalCost {
				t.Log("got:", prd1[i].TotalCost)
				t.Log("exp:", c.TotalCost)
				t.Error("Should match the TotalCost")
			}
		})
	}
}
