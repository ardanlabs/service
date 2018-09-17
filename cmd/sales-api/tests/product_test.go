package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ardanlabs/service/internal/platform/tests"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/ardanlabs/service/internal/product"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gopkg.in/mgo.v2/bson"
)

// TestProducts is the entry point for the products
func TestProducts(t *testing.T) {
	defer tests.Recover(t)

	t.Run("getProducts200Empty", getProducts200Empty)
	t.Run("postProduct400", postProduct400)
	t.Run("getProduct404", getProduct404)
	t.Run("getProduct400", getProduct400)
	t.Run("deleteProduct404", deleteProduct404)
	t.Run("putProduct404", putProduct404)
	t.Run("crudProducts", crudProduct)
}

// getProducts200Empty validates an empty products list can be retrieved with the endpoint.
func getProducts200Empty(t *testing.T) {
	r := httptest.NewRequest("GET", "/v1/products", nil)
	w := httptest.NewRecorder()
	a.ServeHTTP(w, r)

	t.Log("Given the need to fetch an empty list of products with the products endpoint.")
	{
		t.Log("\tTest 0:\tWhen fetching an empty product list.")
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tShould receive a status code of 200 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 200 for the response.", tests.Success)

			recv := w.Body.String()
			resp := `[]`
			if resp != recv {
				t.Log("Got :", recv)
				t.Log("Want:", resp)
				t.Fatalf("\t%s\tShould get the expected result.", tests.Failed)
			}
			t.Logf("\t%s\tShould get the expected result.", tests.Success)
		}
	}
}

// postProduct400 validates a product can't be created with the endpoint
// unless a valid product document is submitted.
func postProduct400(t *testing.T) {
	np := map[string]string{
		"notes": "missing fields",
	}

	body, err := json.Marshal(&np)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest("POST", "/v1/products", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	a.ServeHTTP(w, r)

	t.Log("Given the need to validate a new product can't be created with an invalid document.")
	{
		t.Log("\tTest 0:\tWhen using an incomplete product value.")
		{
			if w.Code != http.StatusBadRequest {
				t.Fatalf("\t%s\tShould receive a status code of 400 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 400 for the response.", tests.Success)

			// Inspect the response.
			var got web.JSONError
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tShould be able to unmarshal the response to an error type : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to unmarshal the response to an error type.", tests.Success)

			// Define what we want to see.
			want := web.JSONError{
				Error: "field validation failure",
				Fields: web.InvalidError{
					{Fld: "Name", Err: "required"},
					{Fld: "Family", Err: "required"},
					{Fld: "UnitPrice", Err: "required"},
					{Fld: "Quantity", Err: "required"},
				},
			}

			// We can't rely on the order of the field errors so they have to be
			// sorted. Tell the cmp package how to sort them.
			sorter := cmpopts.SortSlices(func(a, b web.Invalid) bool {
				return a.Fld < b.Fld
			})

			if diff := cmp.Diff(want, got, sorter); diff != "" {
				t.Fatalf("\t%s\tShould get the expected result. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get the expected result.", tests.Success)
		}
	}
}

// getProduct400 validates a product request for a malformed id.
func getProduct400(t *testing.T) {
	id := "12345"

	r := httptest.NewRequest("GET", "/v1/products/"+id, nil)
	w := httptest.NewRecorder()
	a.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a product with a malformed id.")
	{
		t.Logf("\tTest 0:\tWhen using the new product %s.", id)
		{
			if w.Code != http.StatusBadRequest {
				t.Fatalf("\t%s\tShould receive a status code of 400 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 400 for the response.", tests.Success)

			recv := w.Body.String()
			resp := `{
  "error": "ID is not in its proper form"
}`
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
func getProduct404(t *testing.T) {
	id := bson.NewObjectId().Hex()

	r := httptest.NewRequest("GET", "/v1/products/"+id, nil)
	w := httptest.NewRecorder()
	a.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a product with an unknown id.")
	{
		t.Logf("\tTest 0:\tWhen using the new product %s.", id)
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\tShould receive a status code of 404 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 404 for the response.", tests.Success)

			recv := w.Body.String()
			resp := "Entity not found"
			if !strings.Contains(recv, resp) {
				t.Log("Got :", recv)
				t.Log("Want:", resp)
				t.Fatalf("\t%s\tShould get the expected result.", tests.Failed)
			}
			t.Logf("\t%s\tShould get the expected result.", tests.Success)
		}
	}
}

// deleteProduct404 validates deleting a product that does not exist.
func deleteProduct404(t *testing.T) {
	id := bson.NewObjectId().Hex()

	r := httptest.NewRequest("DELETE", "/v1/products/"+id, nil)
	w := httptest.NewRecorder()
	a.ServeHTTP(w, r)

	t.Log("Given the need to validate deleting a product that does not exist.")
	{
		t.Logf("\tTest 0:\tWhen using the new product %s.", id)
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\tShould receive a status code of 404 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 404 for the response.", tests.Success)

			recv := w.Body.String()
			resp := "Entity not found"
			if !strings.Contains(recv, resp) {
				t.Log("Got :", recv)
				t.Log("Want:", resp)
				t.Fatalf("\t%s\tShould get the expected result.", tests.Failed)
			}
			t.Logf("\t%s\tShould get the expected result.", tests.Success)
		}
	}
}

// putProduct404 validates updating a product that does not exist.
func putProduct404(t *testing.T) {
	up := product.UpdateProduct{
		Name: tests.StringPointer("Nonexistent"),
	}

	id := bson.NewObjectId().Hex()

	body, err := json.Marshal(&up)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest("PUT", "/v1/products/"+id, bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	a.ServeHTTP(w, r)

	t.Log("Given the need to validate updating a product that does not exist.")
	{
		t.Logf("\tTest 0:\tWhen using the new product %s.", id)
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\tShould receive a status code of 404 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 404 for the response.", tests.Success)

			recv := w.Body.String()
			resp := "Entity not found"
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
func crudProduct(t *testing.T) {
	p := postProduct201(t)
	defer deleteProduct204(t, p.ID.Hex())

	getProduct200(t, p.ID.Hex())
	putProduct204(t, p.ID.Hex())
}

// postProduct201 validates a product can be created with the endpoint.
func postProduct201(t *testing.T) product.Product {
	np := product.NewProduct{
		Name:      "Comic Books",
		Notes:     "Various conditions.",
		Family:    "Kennedy",
		UnitPrice: 25,
		Quantity:  60,
	}

	body, err := json.Marshal(&np)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest("POST", "/v1/products", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	a.ServeHTTP(w, r)

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
			want.Notes = "Various conditions."
			want.Family = "Kennedy"
			want.UnitPrice = 25
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
func deleteProduct204(t *testing.T, id string) {
	r := httptest.NewRequest("DELETE", "/v1/products/"+id, nil)
	w := httptest.NewRecorder()
	a.ServeHTTP(w, r)

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
func getProduct200(t *testing.T, id string) {
	r := httptest.NewRequest("GET", "/v1/products/"+id, nil)
	w := httptest.NewRecorder()
	a.ServeHTTP(w, r)

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
			want.ID = bson.ObjectIdHex(id)
			want.Name = "Comic Books"
			want.Notes = "Various conditions."
			want.Family = "Kennedy"
			want.UnitPrice = 25
			want.Quantity = 60

			if diff := cmp.Diff(want, p); diff != "" {
				t.Fatalf("\t%s\tShould get the expected result. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get the expected result.", tests.Success)
		}
	}
}

// putProduct204 validates updating a product that does exist.
func putProduct204(t *testing.T, id string) {
	body := `{"name": "Graphic Novels", "unit_price": 100}`
	r := httptest.NewRequest("PUT", "/v1/products/"+id, strings.NewReader(body))
	w := httptest.NewRecorder()
	a.ServeHTTP(w, r)

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
			a.ServeHTTP(w, r)

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

			if ru.Family != "Kennedy" {
				t.Fatalf("\t%s\tShould not affect other fields like Family : got %q want %q", tests.Failed, ru.Family, "Kennedy")
			}
			t.Logf("\t%s\tShould not affect other fields like Family.", tests.Success)
		}
	}
}
