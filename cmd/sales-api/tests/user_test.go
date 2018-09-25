package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/platform/tests"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/ardanlabs/service/internal/user"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gopkg.in/mgo.v2/bson"
)

// TestUsers is the entry point for testing user management functions.
func TestUsers(t *testing.T) {
	defer tests.Recover(t)

	t.Run("getToken401", getToken401)
	t.Run("getToken200", getToken200)
	t.Run("postUser400", postUser400)
	t.Run("postUser401", postUser401)
	t.Run("postUser403", postUser403)
	t.Run("getUser400", getUser400)
	t.Run("getUser403", getUser403)
	t.Run("getUser404", getUser404)
	t.Run("deleteUser404", deleteUser404)
	t.Run("putUser404", putUser404)
	t.Run("crudUsers", crudUser)
}

// getToken401 ensures an unknown user can't generate a token.
func getToken401(t *testing.T) {
	r := httptest.NewRequest("GET", "/v1/users/token", nil)
	w := httptest.NewRecorder()

	r.SetBasicAuth("unknown@example.com", "some-password")

	a.ServeHTTP(w, r)

	t.Log("Given the need to deny tokens to unknown users.")
	{
		t.Log("\tTest 0:\tWhen fetching a token with an unrecognized email.")
		{
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("\t%s\tShould receive a status code of 401 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 401 for the response.", tests.Success)
		}
	}
}

// getToken200
func getToken200(t *testing.T) {

	r := httptest.NewRequest("GET", "/v1/users/token", nil)
	w := httptest.NewRecorder()

	r.SetBasicAuth("admin@ardanlabs.com", "gophers")

	a.ServeHTTP(w, r)

	t.Log("Given the need to issues tokens to known users.")
	{
		t.Log("\tTest 0:\tWhen fetching a token with valid credentials.")
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tShould receive a status code of 200 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 200 for the response.", tests.Success)

			var got user.Token
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tShould be able to unmarshal the response : %v", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to unmarshal the response.", tests.Success)

			// TODO(jlw) Should we ensure the token is valid?
		}
	}
}

// postUser400 validates a user can't be created with the endpoint
// unless a valid user document is submitted.
func postUser400(t *testing.T) {
	body, err := json.Marshal(&user.User{})
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest("POST", "/v1/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", adminAuthorization)

	a.ServeHTTP(w, r)

	t.Log("Given the need to validate a new user can't be created with an invalid document.")
	{
		t.Log("\tTest 0:\tWhen using an incomplete user value.")
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
					{Fld: "Email", Err: "required"},
					{Fld: "Roles", Err: "required"},
					{Fld: "Password", Err: "required"},
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

// postUser401 validates a user can't be created unless the calling user is
// authenticated.
func postUser401(t *testing.T) {
	body, err := json.Marshal(&user.User{})
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest("POST", "/v1/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", userAuthorization)

	a.ServeHTTP(w, r)

	t.Log("Given the need to validate a new user can't be created with an invalid document.")
	{
		t.Log("\tTest 0:\tWhen using an incomplete user value.")
		{
			if w.Code != http.StatusForbidden {
				t.Fatalf("\t%s\tShould receive a status code of 403 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 403 for the response.", tests.Success)
		}
	}
}

// postUser403 validates a user can't be created unless the calling user is
// an admin user. Regular users can't do this.
func postUser403(t *testing.T) {
	body, err := json.Marshal(&user.User{})
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest("POST", "/v1/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Not setting the Authorization header

	a.ServeHTTP(w, r)

	t.Log("Given the need to validate a new user can't be created with an invalid document.")
	{
		t.Log("\tTest 0:\tWhen using an incomplete user value.")
		{
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("\t%s\tShould receive a status code of 401 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 401 for the response.", tests.Success)
		}
	}
}

// getUser400 validates a user request for a malformed userid.
func getUser400(t *testing.T) {
	id := "12345"

	r := httptest.NewRequest("GET", "/v1/users/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", adminAuthorization)

	a.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a user with a malformed userid.")
	{
		t.Logf("\tTest 0:\tWhen using the new user %s.", id)
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

// getUser403 validates a regular user can't fetch anyone but themselves
func getUser403(t *testing.T) {
	t.Log("Given the need to validate regular users can't fetch other users.")
	{
		t.Logf("\tTest 0:\tWhen fetching the admin user as a regular user.")
		{
			r := httptest.NewRequest("GET", "/v1/users/"+adminID, nil)
			w := httptest.NewRecorder()

			r.Header.Set("Authorization", userAuthorization)

			a.ServeHTTP(w, r)

			if w.Code != http.StatusForbidden {
				t.Fatalf("\t%s\tShould receive a status code of 403 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 403 for the response.", tests.Success)

			recv := w.Body.String()
			resp := "Forbidden"
			if !strings.Contains(recv, resp) {
				t.Log("Got :", recv)
				t.Log("Want:", resp)
				t.Fatalf("\t%s\tShould get the expected result.", tests.Failed)
			}
			t.Logf("\t%s\tShould get the expected result.", tests.Success)
		}

		t.Logf("\tTest 1:\tWhen fetching the user as a themselves.")
		{

			r := httptest.NewRequest("GET", "/v1/users/"+userID, nil)
			w := httptest.NewRecorder()

			r.Header.Set("Authorization", userAuthorization)

			a.ServeHTTP(w, r)
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tShould receive a status code of 200 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 200 for the response.", tests.Success)
		}
	}
}

// getUser404 validates a user request for a user that does not exist with the endpoint.
func getUser404(t *testing.T) {
	id := bson.NewObjectId().Hex()

	r := httptest.NewRequest("GET", "/v1/users/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", adminAuthorization)

	a.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a user with an unknown id.")
	{
		t.Logf("\tTest 0:\tWhen using the new user %s.", id)
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

// deleteUser404 validates deleting a user that does not exist.
func deleteUser404(t *testing.T) {
	id := bson.NewObjectId().Hex()

	r := httptest.NewRequest("DELETE", "/v1/users/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", adminAuthorization)

	a.ServeHTTP(w, r)

	t.Log("Given the need to validate deleting a user that does not exist.")
	{
		t.Logf("\tTest 0:\tWhen using the new user %s.", id)
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

// putUser404 validates updating a user that does not exist.
func putUser404(t *testing.T) {
	u := user.UpdateUser{
		Name: tests.StringPointer("Doesn't Exist"),
	}

	id := bson.NewObjectId().Hex()

	body, err := json.Marshal(&u)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest("PUT", "/v1/users/"+id, bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", adminAuthorization)

	a.ServeHTTP(w, r)

	t.Log("Given the need to validate updating a user that does not exist.")
	{
		t.Logf("\tTest 0:\tWhen using the new user %s.", id)
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

// crudUser performs a complete test of CRUD against the api.
func crudUser(t *testing.T) {
	nu := postUser201(t)
	defer deleteUser204(t, nu.ID.Hex())

	getUser200(t, nu.ID.Hex())
	putUser204(t, nu.ID.Hex())
	putUser403(t, nu.ID.Hex())
}

// postUser201 validates a user can be created with the endpoint.
func postUser201(t *testing.T) user.User {
	nu := user.NewUser{
		Name:            "Bill Kennedy",
		Email:           "bill@ardanlabs.com",
		Roles:           []string{auth.RoleAdmin},
		Password:        "gophers",
		PasswordConfirm: "gophers",
	}

	body, err := json.Marshal(&nu)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest("POST", "/v1/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", adminAuthorization)

	a.ServeHTTP(w, r)

	// u is the value we will return.
	var u user.User

	t.Log("Given the need to create a new user with the users endpoint.")
	{
		t.Log("\tTest 0:\tWhen using the declared user value.")
		{
			if w.Code != http.StatusCreated {
				t.Fatalf("\t%s\tShould receive a status code of 201 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 201 for the response.", tests.Success)

			if err := json.NewDecoder(w.Body).Decode(&u); err != nil {
				t.Fatalf("\t%s\tShould be able to unmarshal the response : %v", tests.Failed, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like ID and Dates so we copy u.
			want := u
			want.Name = "Bill Kennedy"
			want.Email = "bill@ardanlabs.com"
			want.Roles = []string{auth.RoleAdmin}

			if diff := cmp.Diff(want, u); diff != "" {
				t.Fatalf("\t%s\tShould get the expected result. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get the expected result.", tests.Success)
		}
	}

	return u
}

// deleteUser200 validates deleting a user that does exist.
func deleteUser204(t *testing.T, id string) {
	r := httptest.NewRequest("DELETE", "/v1/users/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", adminAuthorization)

	a.ServeHTTP(w, r)

	t.Log("Given the need to validate deleting a user that does exist.")
	{
		t.Logf("\tTest 0:\tWhen using the new user %s.", id)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tShould receive a status code of 204 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 204 for the response.", tests.Success)
		}
	}
}

// getUser200 validates a user request for an existing userid.
func getUser200(t *testing.T, id string) {
	r := httptest.NewRequest("GET", "/v1/users/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", adminAuthorization)

	a.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a user that exsits.")
	{
		t.Logf("\tTest 0:\tWhen using the new user %s.", id)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tShould receive a status code of 200 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 200 for the response.", tests.Success)

			var u user.User
			if err := json.NewDecoder(w.Body).Decode(&u); err != nil {
				t.Fatalf("\t%s\tShould be able to unmarshal the response : %v", tests.Failed, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like Dates so we copy p.
			want := u
			want.ID = bson.ObjectIdHex(id)
			want.Name = "Bill Kennedy"
			want.Email = "bill@ardanlabs.com"
			want.Roles = []string{auth.RoleAdmin}

			if diff := cmp.Diff(want, u); diff != "" {
				t.Fatalf("\t%s\tShould get the expected result. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get the expected result.", tests.Success)
		}
	}
}

// putUser204 validates updating a user that does exist.
func putUser204(t *testing.T, id string) {
	body := `{"name": "Jacob Walker"}`

	r := httptest.NewRequest("PUT", "/v1/users/"+id, strings.NewReader(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", adminAuthorization)

	a.ServeHTTP(w, r)

	t.Log("Given the need to update a user with the users endpoint.")
	{
		t.Log("\tTest 0:\tWhen using the modified user value.")
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tShould receive a status code of 204 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 204 for the response.", tests.Success)

			r = httptest.NewRequest("GET", "/v1/users/"+id, nil)
			w = httptest.NewRecorder()

			r.Header.Set("Authorization", adminAuthorization)

			a.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tShould receive a status code of 200 for the retrieve : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 200 for the retrieve.", tests.Success)

			var ru user.User
			if err := json.NewDecoder(w.Body).Decode(&ru); err != nil {
				t.Fatalf("\t%s\tShould be able to unmarshal the response : %v", tests.Failed, err)
			}

			if ru.Name != "Jacob Walker" {
				t.Fatalf("\t%s\tShould see an updated Name : got %q want %q", tests.Failed, ru.Name, "Jacob Walker")
			}
			t.Logf("\t%s\tShould see an updated Name.", tests.Success)

			if ru.Email != "bill@ardanlabs.com" {
				t.Fatalf("\t%s\tShould not affect other fields like Email : got %q want %q", tests.Failed, ru.Email, "bill@ardanlabs.com")
			}
			t.Logf("\t%s\tShould not affect other fields like Email.", tests.Success)
		}
	}
}

// putUser403 validates that a user can't modify users unless they are an admin.
func putUser403(t *testing.T, id string) {
	body := `{"name": "Anna Walker"}`

	r := httptest.NewRequest("PUT", "/v1/users/"+id, strings.NewReader(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", userAuthorization)

	a.ServeHTTP(w, r)

	t.Log("Given the need to update a user with the users endpoint.")
	{
		t.Log("\tTest 0:\tWhen a non-admin user makes a request")
		{
			if w.Code != http.StatusForbidden {
				t.Fatalf("\t%s\tShould receive a status code of 403 for the response : %v", tests.Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 403 for the response.", tests.Success)
		}
	}
}
