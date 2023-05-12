package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/ardanlabs/service/app/services/sales-api/handlers"
	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/productgrp"
	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/sys/validate"
	v1Web "github.com/ardanlabs/service/business/web/v1"
	"github.com/ardanlabs/service/business/web/v1/paging"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
)

// ProductTests holds methods for each product subtest. This type allows
// passing dependencies for tests while still providing a convenient syntax
// when subtests are registered.
type ProductTests struct {
	app       http.Handler
	userToken string
}

// Test_Products is the entry point for testing product apis.
func Test_Products(t *testing.T) {
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
	tests := ProductTests{
		app: handlers.APIMux(handlers.APIMuxConfig{
			Shutdown: shutdown,
			Log:      test.Log,
			Auth:     test.Auth,
			DB:       test.DB,
		}),
		userToken: test.Token("admin@example.com", "gophers"),
	}

	// -------------------------------------------------------------------------

	seed := func(ctx context.Context, usrCore *user.Core, prdCore *product.Core) ([]product.Product, error) {
		var filter user.QueryFilter
		filter.WithName("Admin Gopher")

		usrs, err := usrCore.Query(ctx, filter, user.DefaultOrderBy, 1, 1)
		if err != nil {
			return nil, fmt.Errorf("seeding users : %w", err)
		}

		prds, err := prdCore.GenerateFakeProducts(10, usrs[0].ID)
		if err != nil {
			return nil, fmt.Errorf("seeding products : %w", err)
		}

		return prds, nil
	}

	t.Log("Go seeding ...")

	prds, err := seed(context.Background(), api.User, api.Product)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	t.Run("postProduct400", tests.postProduct400())
	t.Run("postProduct401", tests.postProduct401())
	t.Run("getProduct404", tests.getProduct404())
	t.Run("getProduct400", tests.getProduct400())
	t.Run("deleteProductNotFound", tests.deleteProductNotFound())
	t.Run("putProduct404", tests.putProduct404())
	t.Run("crudProducts", tests.crudProduct())
	t.Run("getProducts200", tests.getProducts200(prds))
}

func (pt *ProductTests) postProduct400() func(t *testing.T) {
	return func(t *testing.T) {
		url := "/v1/products"

		r := httptest.NewRequest(http.MethodPost, url, strings.NewReader(`{}`))
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+pt.userToken)
		pt.app.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("Should receive a status code of 400 for the response : %d", w.Code)
		}

		var got v1Web.ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("Should be able to unmarshal the response to an error type : %s", err)
		}

		fields := validate.FieldErrors{
			{Field: "name", Err: "name is a required field"},
			{Field: "cost", Err: "cost is a required field"},
			{Field: "quantity", Err: "quantity must be 1 or greater"},
			{Field: "userID", Err: "userID is a required field"},
		}
		exp := v1Web.ErrorResponse{
			Error:  "data validation error",
			Fields: fields.Fields(),
		}

		// We can't rely on the order of the field errors so they have to be
		// sorted. Tell the cmp package how to sort them.
		sorter := cmpopts.SortSlices(func(a, b validate.FieldError) bool {
			return a.Field < b.Field
		})

		if diff := cmp.Diff(got, exp, sorter); diff != "" {
			t.Fatalf("Should get the expected result, Diff:\n%s", diff)
		}
	}
}

func (pt *ProductTests) postProduct401() func(t *testing.T) {
	return func(t *testing.T) {
		np := product.NewProduct{
			Name:     "Comic Books",
			Cost:     25,
			Quantity: 60,
		}

		body, err := json.Marshal(&np)
		if err != nil {
			t.Fatal(err)
		}

		url := "/v1/products"

		r := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		pt.app.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Should receive a status code of 401 for the response : %d", w.Code)
		}
	}
}

func (pt *ProductTests) getProduct400() func(t *testing.T) {
	return func(t *testing.T) {
		url := fmt.Sprintf("/v1/products/%d", 12345)

		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+pt.userToken)
		pt.app.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("Should receive a status code of 400 for the response : %d", w.Code)
		}

		got := w.Body.String()
		exp := `{"error":"data validation error","fields":{"product_id":"invalid UUID length: 5"}}`
		if got != exp {
			t.Logf("got: %v", got)
			t.Logf("exp: %v", exp)
			t.Error("Should get the expected result")
		}
	}
}

func (pt *ProductTests) getProduct404() func(t *testing.T) {
	return func(t *testing.T) {
		url := fmt.Sprintf("/v1/products/%s", "a224a8d6-3f9e-4b11-9900-e81a25d80702")

		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+pt.userToken)
		pt.app.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Fatalf("Should receive a status code of 404 for the response : %d", w.Code)
		}

		got := w.Body.String()
		exp := "not found"
		if !strings.Contains(got, exp) {
			t.Logf("got: %v", got)
			t.Logf("exp: %v", exp)
			t.Error("Should get the expected result")
		}
	}
}

func (pt *ProductTests) deleteProductNotFound() func(t *testing.T) {
	return func(t *testing.T) {
		url := fmt.Sprintf("/v1/products/%s", "112262f1-1a77-4374-9f22-39e575aa6348")

		r := httptest.NewRequest(http.MethodDelete, url, nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+pt.userToken)
		pt.app.ServeHTTP(w, r)

		if w.Code != http.StatusNoContent {
			t.Fatalf("Should receive a status code of 204 for the response : %d", w.Code)
		}
	}
}

func (pt *ProductTests) putProduct404() func(t *testing.T) {
	return func(t *testing.T) {
		up := product.UpdateProduct{
			Name: dbtest.StringPointer("Nonexistent"),
		}
		body, err := json.Marshal(&up)
		if err != nil {
			t.Fatal(err)
		}

		url := fmt.Sprintf("/v1/products/%s", "9b468f90-1cf1-4377-b3fa-68b450d632a0")

		r := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+pt.userToken)
		pt.app.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Fatalf("Should receive a status code of 404 for the response : %d", w.Code)
		}

		got := w.Body.String()
		exp := "not found"
		if !strings.Contains(got, exp) {
			t.Logf("got: %v", got)
			t.Logf("exp: %v", exp)
			t.Error("Should get the expected result")
		}
	}
}

// getProducts200 validates a query request.
func (pt *ProductTests) getProducts200(prds []product.Product) func(t *testing.T) {
	return func(t *testing.T) {
		url := "/v1/products?page=1&rows=2"

		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+pt.userToken)
		pt.app.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("Should receive a status code of 200 for the response : %d", w.Code)
		}

		var pr paging.Response[productgrp.AppProduct]
		if err := json.Unmarshal(w.Body.Bytes(), &pr); err != nil {
			t.Fatalf("Should be able to unmarshal the response : %s", err)
		}

		if pr.Total != len(prds) {
			t.Log("got:", pr.Total)
			t.Log("exp:", len(prds))
			t.Error("Should get the right total")
		}

		if len(pr.Items) != 2 {
			t.Log("got:", len(pr.Items))
			t.Log("exp:", 2)
			t.Error("Should get the right number of products")
		}
	}
}

func (pt *ProductTests) crudProduct() func(t *testing.T) {
	return func(t *testing.T) {
		prd := pt.postProduct201(t)
		defer pt.deleteProduct204(t, prd.ID)

		pt.getProduct200(t, prd.ID)
		pt.putProduct200(t, prd.ID)
	}
}

// postProduct201 validates a product can be created with the endpoint.
func (pt *ProductTests) postProduct201(t *testing.T) product.Product {
	np := product.NewProduct{
		Name:     "Comic Books",
		Cost:     25,
		Quantity: 60,
		UserID:   uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
	}

	body, err := json.Marshal(&np)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/products", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)
	pt.app.ServeHTTP(w, r)

	var newPrd product.Product

	if w.Code != http.StatusCreated {
		t.Fatalf("Should receive a status code of 201 for the response : %d", w.Code)
	}

	if err := json.NewDecoder(w.Body).Decode(&newPrd); err != nil {
		t.Fatalf("Should be able to unmarshal the response : %s", err)
	}

	exp := newPrd
	exp.Name = "Comic Books"
	exp.Cost = 25
	exp.Quantity = 60

	if diff := cmp.Diff(newPrd, exp); diff != "" {
		t.Fatalf("Should get the expected result, Diff:\n%s", diff)
	}

	return newPrd
}

// deleteProduct200 validates deleting a product that does exist.
func (pt *ProductTests) deleteProduct204(t *testing.T, id uuid.UUID) {
	url := fmt.Sprintf("/v1/products/%s", id)

	r := httptest.NewRequest(http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)
	pt.app.ServeHTTP(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Should receive a status code of 204 for the response : %d", w.Code)
	}
}

func (pt *ProductTests) getProduct200(t *testing.T, id uuid.UUID) {
	url := fmt.Sprintf("/v1/products/%s", id)

	r := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)
	pt.app.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("Should receive a status code of 200 for the response : %d", w.Code)
	}

	var got product.Product
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("Should be able to unmarshal the response : %s", err)
	}

	exp := got
	exp.ID = id
	exp.Name = "Comic Books"
	exp.Cost = 25
	exp.Quantity = 60

	if diff := cmp.Diff(got, exp); diff != "" {
		t.Fatalf("Should get the expected result, Diff:\n%s", diff)
	}
}

// putProduct200 validates updating a product that does exist.
func (pt *ProductTests) putProduct200(t *testing.T, id uuid.UUID) {
	body := `{"name": "Graphic Novels", "cost": 100}`

	url := fmt.Sprintf("/v1/products/%s", id)

	r := httptest.NewRequest(http.MethodPut, url, strings.NewReader(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)
	pt.app.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("Should receive a status code of 204 for the response : %d", w.Code)
	}

	r = httptest.NewRequest(http.MethodGet, "/v1/products/"+id.String(), nil)
	w = httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)
	pt.app.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("Should receive a status code of 200 for the retrieve : %d", w.Code)
	}

	var ru product.Product
	if err := json.NewDecoder(w.Body).Decode(&ru); err != nil {
		t.Fatalf("Should be able to unmarshal the response : %s", err)
	}

	if ru.Name != "Graphic Novels" {
		t.Fatalf("Should see an updated Name : got %q want %q", ru.Name, "Graphic Novels")
	}
}
