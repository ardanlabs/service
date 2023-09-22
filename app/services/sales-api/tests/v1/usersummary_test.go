package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime/debug"
	"testing"

	v1 "github.com/ardanlabs/service/app/services/sales-api/v1"
	"github.com/ardanlabs/service/app/services/sales-api/v1/cmd/all"
	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/usersummarygrp"
	"github.com/ardanlabs/service/app/services/sales-api/v1/paging"
	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/data/order"
)

// UserTests holds methods for each user subtest. This type allows passing
// dependencies for tests while still providing a convenient syntax when
// subtests are registered.
type UserSummaryTests struct {
	app        http.Handler
	userToken  string
	adminToken string
}

// Test_UserSummary is the entry point for testing user management apis.
func Test_UserSummary(t *testing.T) {
	t.Parallel()

	test := dbtest.NewTest(t, c)
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.Teardown()
	}()

	api := test.CoreAPIs

	shutdown := make(chan os.Signal, 1)
	tests := UserSummaryTests{
		app: v1.APIMux(v1.APIMuxConfig{
			Shutdown: shutdown,
			Log:      test.Log,
			Auth:     test.Auth,
			DB:       test.DB,
		}, all.Routes()),
		userToken:  test.Token("user@example.com", "gophers"),
		adminToken: test.Token("admin@example.com", "gophers"),
	}

	// -------------------------------------------------------------------------

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

	t.Log("Go seeding ...")

	usrs, _, err := seed(context.Background(), api.User, api.Product)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	t.Run("query", tests.query(usrs))
}

func (ust *UserSummaryTests) query(usrs []user.User) func(t *testing.T) {
	return func(t *testing.T) {
		url := "/v1/usersummary"

		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ust.adminToken)
		ust.app.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("Should receive a status code of 200 for the response : %d", w.Code)
		}

		var pr paging.Response[usersummarygrp.AppUserSummary]
		if err := json.Unmarshal(w.Body.Bytes(), &pr); err != nil {
			t.Fatalf("Should be able to unmarshal the response : %s", err)
		}

		if pr.Total != len(usrs) {
			t.Log("got:", pr.Total)
			t.Log("exp:", len(usrs))
			t.Error("Should get the right total")
		}

		if len(pr.Items) != len(usrs) {
			t.Log("got:", len(pr.Items))
			t.Log("exp:", len(usrs))
			t.Error("Should get the right number of users")
		}
	}
}
