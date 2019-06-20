package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/ardanlabs/service/cmd/sales-api/handlers"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/ardanlabs/service/internal/product"
	"github.com/ardanlabs/service/internal/tests"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// TestProducts runs a series of tests to exercise Product behavior from the
// API level. The subtests all share the same database and application for
// speed and convenience. The downside is the order the tests are ran matters
// and one test may break if other tests are not ran before it. If a particular
// subtest needs a fresh instance of the application it can make it or it
// should be its own Test* function.
func TestProducts(t *testing.T) {
	test := tests.NewIntegration(t)
	defer test.Teardown()

	shutdown := make(chan os.Signal, 1)
	tests := ProductTests{
		app:       handlers.API(shutdown, test.Log, test.DB, test.Authenticator),
		userToken: test.Token("admin@example.com", "gophers"),
	}

	t.Run("postProduct400", tests.postProduct400)
	t.Run("postProduct401", tests.postProduct401)
	t.Run("getProduct404", tests.getProduct404)
	t.Run("getProduct400", tests.getProduct400)
	t.Run("deleteProductNotFound", tests.deleteProductNotFound)
	t.Run("putProduct404", tests.putProduct404)
	t.Run("crudProducts", tests.crudProduct)
}

// ProductTests holds methods for each product subtest. This type allows
// passing dependencies for tests while still providing a convenient syntax
// when subtests are registered.
type ProductTests struct {
	app       http.Handler
	userToken string
}

// postProduct400 validates a product can't be created with the endpoint
// unless a valid product document is submitted.
func (pt *ProductTests) postProduct400(t *testing.T) {
	r := httptest.NewRequest("POST", "/v1/products", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new product can't be created with an invalid document.")
	{
		t.Log("\tTest 0:\tWhen using an incomplete product value.")
		{
			if w.Code != http.StatusBadRequest {
				t.Fatalf("\t%s\tShould receive a status code of 400 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 400 for the response.", tests.Success)

			// Inspect the response.
			var got web.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tShould be able to unmarshal the response to an error type : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to unmarshal the response to an error type.", tests.Success)

			// Define what we want to see.
			want := web.ErrorResponse{
				Error: "field validation error",
				Fields: []web.FieldError{
					{Field: "name", Error: "name is a required field"},
					{Field: "cost", Error: "cost is a required field"},
					{Field: "quantity", Error: "quantity must be 1 or greater"},
				},
			}

			// We can't rely on the order of the field errors so they have to be
			// sorted. Tell the cmp package how to sort them.
			sorter := cmpopts.SortSlices(func(a, b web.FieldError) bool {
				return a.Field < b.Field
			})

			if diff := cmp.Diff(want, got, sorter); diff != "" {
				t.Fatalf("\t%s\tShould get the expected result. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get the expected result.", tests.Success)
		}
	}
}

// postProduct401 validates a product can't be created with the endpoint
// unless the user is authenticated
func (pt *ProductTests) postProduct401(t *testing.T) {
	np := product.NewProduct{
		Name:     "Comic Books",
		Cost:     25,
		Quantity: 60,
	}

	body, err := json.Marshal(&np)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest("POST", "/v1/products", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Not setting an authorization header

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new product can't be created with an invalid document.")
	{
		t.Log("\tTest 0:\tWhen using an incomplete product value.")
		{
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("\t%s\tShould receive a status code of 401 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 401 for the response.", tests.Success)
		}
	}
}

// getProduct400 validates a product request for a malformed id.
func (pt *ProductTests) getProduct400(t *testing.T) {
	id := "12345"

	r := httptest.NewRequest("GET", "/v1/products/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a product with a malformed id.")
	{
		t.Logf("\tTest 0:\tWhen using the new product %s.", id)
		{
			if w.Code != http.StatusBadRequest {
				t.Fatalf("\t%s\tShould receive a status code of 400 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 400 for the response.", tests.Success)

			recv := w.Body.String()
			resp := `{"error":"ID is not in its proper form"}`
			if resp != recv {
				t.Log("Got :", recv)
				t.Log("Want:", resp)
				t.Fatalf("\t%s\tShould get the expected result.", tests.Failed)
			}
			t.Logf("\t%s\tShould get the expected result.", tests.Success)
		}
	}
}

// getProduct404 validates a product request for a product that does not exist with the endpoint.
func (pt *ProductTests) getProduct404(t *testing.T) {
	id := "a224a8d6-3f9e-4b11-9900-e81a25d80702"

	r := httptest.NewRequest("GET", "/v1/products/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a product with an unknown id.")
	{
		t.Logf("\tTest 0:\tWhen using the new product %s.", id)
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\tShould receive a status code of 404 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 404 for the response.", tests.Success)

			recv := w.Body.String()
			resp := "Product not found"
			if !strings.Contains(recv, resp) {
				t.Log("Got :", recv)
				t.Log("Want:", resp)
				t.Fatalf("\t%s\tShould get the expected result.", tests.Failed)
			}
			t.Logf("\t%s\tShould get the expected result.", tests.Success)
		}
	}
}

// deleteProductNotFound validates deleting a product that does not exist is not a failure.
func (pt *ProductTests) deleteProductNotFound(t *testing.T) {
	id := "112262f1-1a77-4374-9f22-39e575aa6348"

	r := httptest.NewRequest("DELETE", "/v1/products/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate deleting a product that does not exist.")
	{
		t.Logf("\tTest 0:\tWhen using the new product %s.", id)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tShould receive a status code of 204 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 204 for the response.", tests.Success)
		}
	}
}

// putProduct404 validates updating a product that does not exist.
func (pt *ProductTests) putProduct404(t *testing.T) {
	up := product.UpdateProduct{
		Name: tests.StringPointer("Nonexistent"),
	}

	id := "9b468f90-1cf1-4377-b3fa-68b450d632a0"

	body, err := json.Marshal(&up)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest("PUT", "/v1/products/"+id, bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate updating a product that does not exist.")
	{
		t.Logf("\tTest 0:\tWhen using the new product %s.", id)
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\tShould receive a status code of 404 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 404 for the response.", tests.Success)

			recv := w.Body.String()
			resp := "Product not found"
			if !strings.Contains(recv, resp) {
				t.Log("Got :", recv)
				t.Log("Want:", resp)
				t.Fatalf("\t%s\tShould get the expected result.", tests.Failed)
			}
			t.Logf("\t%s\tShould get the expected result.", tests.Success)
		}
	}
}

// crudProduct performs a complete test of CRUD against the api.
func (pt *ProductTests) crudProduct(t *testing.T) {
	p := pt.postProduct201(t)
	defer pt.deleteProduct204(t, p.ID)

	pt.getProduct200(t, p.ID)
	pt.putProduct204(t, p.ID)
}

// postProduct201 validates a product can be created with the endpoint.
func (pt *ProductTests) postProduct201(t *testing.T) product.Product {
	np := product.NewProduct{
		Name:     "Comic Books",
		Cost:     25,
		Quantity: 60,
	}

	body, err := json.Marshal(&np)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest("POST", "/v1/products", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	// p is the value we will return.
	var p product.Product

	t.Log("Given the need to create a new product with the products endpoint.")
	{
		t.Log("\tTest 0:\tWhen using the declared product value.")
		{
			if w.Code != http.StatusCreated {
				t.Fatalf("\t%s\tShould receive a status code of 201 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 201 for the response.", tests.Success)

			if err := json.NewDecoder(w.Body).Decode(&p); err != nil {
				t.Fatalf("\t%s\tShould be able to unmarshal the response : %v", tests.Failed, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like ID and Dates so we copy p.
			want := p
			want.Name = "Comic Books"
			want.Cost = 25
			want.Quantity = 60

			if diff := cmp.Diff(want, p); diff != "" {
				t.Fatalf("\t%s\tShould get the expected result. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get the expected result.", tests.Success)
		}
	}

	return p
}

// deleteProduct200 validates deleting a product that does exist.
func (pt *ProductTests) deleteProduct204(t *testing.T, id string) {
	r := httptest.NewRequest("DELETE", "/v1/products/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate deleting a product that does exist.")
	{
		t.Logf("\tTest 0:\tWhen using the new product %s.", id)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tShould receive a status code of 204 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 204 for the response.", tests.Success)
		}
	}
}

// getProduct200 validates a product request for an existing id.
func (pt *ProductTests) getProduct200(t *testing.T, id string) {
	r := httptest.NewRequest("GET", "/v1/products/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a product that exists.")
	{
		t.Logf("\tTest 0:\tWhen using the new product %s.", id)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tShould receive a status code of 200 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 200 for the response.", tests.Success)

			var p product.Product
			if err := json.NewDecoder(w.Body).Decode(&p); err != nil {
				t.Fatalf("\t%s\tShould be able to unmarshal the response : %v", tests.Failed, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like Dates so we copy p.
			want := p
			want.ID = id
			want.Name = "Comic Books"
			want.Cost = 25
			want.Quantity = 60

			if diff := cmp.Diff(want, p); diff != "" {
				t.Fatalf("\t%s\tShould get the expected result. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get the expected result.", tests.Success)
		}
	}
}

// putProduct204 validates updating a product that does exist.
func (pt *ProductTests) putProduct204(t *testing.T, id string) {
	body := `{"name": "Graphic Novels", "cost": 100}`
	r := httptest.NewRequest("PUT", "/v1/products/"+id, strings.NewReader(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to update a product with the products endpoint.")
	{
		t.Log("\tTest 0:\tWhen using the modified product value.")
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tShould receive a status code of 204 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 204 for the response.", tests.Success)

			r = httptest.NewRequest("GET", "/v1/products/"+id, nil)
			w = httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+pt.userToken)

			pt.app.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tShould receive a status code of 200 for the retrieve : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 200 for the retrieve.", tests.Success)

			var ru product.Product
			if err := json.NewDecoder(w.Body).Decode(&ru); err != nil {
				t.Fatalf("\t%s\tShould be able to unmarshal the response : %v", tests.Failed, err)
			}

			if ru.Name != "Graphic Novels" {
				t.Fatalf("\t%s\tShould see an updated Name : got %q want %q", tests.Failed, ru.Name, "Graphic Novels")
			}
			t.Logf("\t%s\tShould see an updated Name.", tests.Success)
		}
	}
}
