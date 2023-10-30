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

	"github.com/ardanlabs/service/app/services/sales-api/v1/cmd/all"
	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/homegrp"
	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/productgrp"
	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/usergrp"
	"github.com/ardanlabs/service/business/core/home"
	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/data/order"
	v1 "github.com/ardanlabs/service/business/web/v1"
	"github.com/ardanlabs/service/business/web/v1/response"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func Test_Query(t *testing.T) {
	t.Parallel()

	test := dbtest.NewTest(t, c)
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.Teardown()
	}()

	tests := QueryTests{
		app: v1.APIMux(v1.APIMuxConfig{
			Shutdown: make(chan os.Signal, 1),
			Log:      test.Log,
			Auth:     test.V1.Auth,
			DB:       test.DB,
		}, all.Routes()),
		userToken:  test.TokenV1("user@example.com", "gophers"),
		adminToken: test.TokenV1("admin@example.com", "gophers"),
	}

	// -------------------------------------------------------------------------

	t.Log("Seeding data ...")
	sd, err := querySeed(context.Background(), test.CoreAPIs)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	tests.query200(t, sd)
	tests.queryByID200(t, sd)
}

func querySeed(ctx context.Context, api dbtest.CoreAPIs) (seedData, error) {
	usrs, err := api.User.Query(ctx, user.QueryFilter{}, order.By{Field: user.OrderByName, Direction: order.ASC}, 1, 2)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users : %w", err)
	}

	prds1, err := product.TestGenerateSeedProducts(5, api.Product, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding products : %w", err)
	}

	prds2, err := product.TestGenerateSeedProducts(5, api.Product, usrs[1].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding products : %w", err)
	}

	var prds []product.Product
	prds = append(prds, prds1...)
	prds = append(prds, prds2...)

	hmes1, err := home.TestGenerateSeedHomes(5, api.Home, usrs[0].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding homes1 : %w", err)
	}

	hmes2, err := home.TestGenerateSeedHomes(5, api.Home, usrs[1].ID)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding homes2 : %w", err)
	}

	var hmes []home.Home
	hmes = append(hmes, hmes1...)
	hmes = append(hmes, hmes2...)

	sd := seedData{
		users:    usrs,
		products: prds,
		homes:    hmes,
	}

	return sd, nil
}

// =============================================================================

type QueryTests struct {
	app        http.Handler
	userToken  string
	adminToken string
}

func (qt *QueryTests) query200(t *testing.T, sd seedData) {
	usrs := make(map[uuid.UUID]user.User)
	for _, usr := range sd.users {
		usrs[usr.ID] = usr
	}

	table := []struct {
		name    string
		url     string
		resp    any
		expResp any
	}{
		{
			name: "user",
			url:  "/v1/users?page=1&rows=2&orderBy=user_id,DESC",
			resp: &response.PageDocument[usergrp.AppUser]{},
			expResp: &response.PageDocument[usergrp.AppUser]{
				Page:        1,
				RowsPerPage: 2,
				Total:       len(sd.users),
				Items:       toAppUsers(sd.users),
			},
		},
		{
			name: "product",
			url:  "/v1/products?page=1&rows=10&orderBy=user_id,DESC",
			resp: &response.PageDocument[productgrp.AppProductDetails]{},
			expResp: &response.PageDocument[productgrp.AppProductDetails]{
				Page:        1,
				RowsPerPage: 10,
				Total:       len(sd.products),
				Items:       toAppProductsDetails(sd.products, usrs),
			},
		},
		{
			name: "home",
			url:  "/v1/homes?page=1&rows=10&orderBy=user_id,DESC",
			resp: &response.PageDocument[homegrp.AppHome]{},
			expResp: &response.PageDocument[homegrp.AppHome]{
				Page:        1,
				RowsPerPage: 10,
				Total:       len(sd.products),
				Items:       toAppHomes(sd.homes),
			},
		},
	}

	for _, tt := range table {
		f := func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+qt.adminToken)
			qt.app.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("%s: Should receive a status code of 200 for the response : %d", tt.name, w.Code)
			}

			if err := json.Unmarshal(w.Body.Bytes(), tt.resp); err != nil {
				t.Fatalf("Should be able to unmarshal the response : %s", err)
			}

			diff := cmp.Diff(tt.resp, tt.expResp)
			if diff != "" {
				t.Log("GOT")
				t.Logf("%#v", tt.resp)
				t.Log("EXP")
				t.Logf("%#v", tt.expResp)
				t.Fatalf("Should get the expected response")
			}
		}

		t.Run("query200-"+tt.name, f)
	}
}

func (qt *QueryTests) queryByID200(t *testing.T, sd seedData) {
	table := []struct {
		name    string
		url     string
		resp    any
		expResp any
	}{
		{
			name:    "user",
			url:     fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
			resp:    &usergrp.AppUser{},
			expResp: toAppUserPtr(sd.users[0]),
		},
		{
			name:    "product",
			url:     fmt.Sprintf("/v1/products/%s", sd.products[0].ID),
			resp:    &productgrp.AppProduct{},
			expResp: toAppProductPtr(sd.products[0]),
		},
		{
			name:    "home",
			url:     fmt.Sprintf("/v1/homes/%s", sd.homes[0].ID),
			resp:    &homegrp.AppHome{},
			expResp: toAppHomePtr(sd.homes[0]),
		},
	}

	for _, tt := range table {
		f := func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+qt.adminToken)
			qt.app.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("%s: Should receive a status code of 200 for the response : %d", tt.name, w.Code)
			}

			if err := json.Unmarshal(w.Body.Bytes(), tt.resp); err != nil {
				t.Fatalf("Should be able to unmarshal the response : %s", err)
			}

			diff := cmp.Diff(tt.resp, tt.expResp)
			if diff != "" {
				t.Log("GOT")
				t.Logf("%#v", tt.resp)
				t.Log("EXP")
				t.Logf("%#v", tt.expResp)
				t.Fatalf("Should get the expected response")
			}
		}

		t.Run("queryByID200-"+tt.name, f)
	}
}
