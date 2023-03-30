package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ardanlabs/service/app/services/sales-api/handlers"
	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/usergrp"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/sys/validate"
	v1Web "github.com/ardanlabs/service/business/web/v1"
	"github.com/ardanlabs/service/business/web/v1/paging"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
)

// UserTests holds methods for each user subtest. This type allows passing
// dependencies for tests while still providing a convenient syntax when
// subtests are registered.
type UserTests struct {
	app        http.Handler
	userToken  string
	adminToken string
}

// Test_Users is the entry point for testing user management functions.
func Test_Users(t *testing.T) {
	t.Parallel()

	test := dbtest.NewIntegration(t, c)
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
		test.Teardown()
	}()

	shutdown := make(chan os.Signal, 1)
	tests := UserTests{
		app: handlers.APIMux(handlers.APIMuxConfig{
			Shutdown: shutdown,
			Log:      test.Log,
			Auth:     test.Auth,
			DB:       test.DB,
		}),
		userToken:  test.Token("user@example.com", "gophers"),
		adminToken: test.Token("admin@example.com", "gophers"),
	}

	t.Run("getToken404", tests.getToken404)
	t.Run("getToken200", tests.getToken200)
	t.Run("postUser400", tests.postUser400)
	t.Run("postUser401", tests.postUser401)
	t.Run("postAdmin401", tests.postAdmin401)
	t.Run("getUser400", tests.getUser400)
	t.Run("getUser401", tests.getUser401)
	t.Run("getUser404", tests.getUser404)
	t.Run("deleteUserNotFound", tests.deleteUserNotFound)
	t.Run("putUser404", tests.putUser404)
	t.Run("crudUsers", tests.crudUser)
	t.Run("getUsers200", tests.getUsers200)
}

// getToken401 ensures an unknown user can't generate a token.
func (ut *UserTests) getToken404(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/users/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1", nil)
	w := httptest.NewRecorder()

	r.SetBasicAuth("unknown@example.com", "some-password")
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to deny tokens to unknown users.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen fetching a token with an unrecognized email.", testID)
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 404 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 404 for the response.", dbtest.Success, testID)
		}
	}
}

func (ut *UserTests) getToken200(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/users/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1", nil)
	w := httptest.NewRecorder()

	r.SetBasicAuth("admin@example.com", "gophers")
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to issues tokens to known users.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen fetching a token with valid credentials.", testID)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", dbtest.Success, testID)

			var got struct {
				Token string `json:"token"`
			}
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to unmarshal the response.", dbtest.Success, testID)

			// TODO(jlw) Should we ensure the token is valid?
		}
	}
}

// postUser400 validates a user can't be created with the endpoint
// unless a valid user document is submitted.
func (ut *UserTests) postUser400(t *testing.T) {
	usr := usergrp.AppNewUser{
		Email: "bill@ardanlabs.com",
	}

	body, err := json.Marshal(usr)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new user can't be created with an invalid document.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using an incomplete user value.", testID)
		{
			if w.Code != http.StatusBadRequest {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 400 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 400 for the response.", dbtest.Success, testID)

			var got v1Web.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response to an error type : %v", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to unmarshal the response to an error type.", dbtest.Success, testID)

			fields := validate.FieldErrors{
				{Field: "name", Err: "name is a required field"},
				{Field: "roles", Err: "roles is a required field"},
				{Field: "password", Err: "password is a required field"},
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
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)
		}
	}
}

// postAdmin401 validates a user can't be created unless the calling user is
// an admin. Regular users can't do this.
func (ut *UserTests) postAdmin401(t *testing.T) {
	body, err := json.Marshal(&usergrp.AppNewUser{})
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.userToken)
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new user can't be created with an invalid document.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using an incomplete user value.", testID)
		{
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 401 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 401 for the response.", dbtest.Success, testID)
		}
	}
}

// postUser401 validates a user can't be created unless the calling user is
// authenticated.
func (ut *UserTests) postUser401(t *testing.T) {
	body, err := json.Marshal(&usergrp.AppNewUser{})
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Not setting the Authorization header.
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new user can't be created with an invalid document.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using an incomplete user value.", testID)
		{
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 401 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 401 for the response.", dbtest.Success, testID)
		}
	}
}

// getUser400 validates a user request for a malformed userid.
func (ut *UserTests) getUser400(t *testing.T) {
	id := "12345"

	r := httptest.NewRequest(http.MethodGet, "/v1/users/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a user with a malformed userid.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new user %s.", testID, id)
		{
			if w.Code != http.StatusBadRequest {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 400 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 400 for the response.", dbtest.Success, testID)

			got := w.Body.String()
			exp := `{"error":"ID is not in its proper form"}`
			if got != exp {
				t.Logf("\t\tTest %d:\tGot : %v", testID, got)
				t.Logf("\t\tTest %d:\tExp: %v", testID, exp)
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)
		}
	}
}

// getUser401 validates a regular user can't fetch anyone but themselves.
func (ut *UserTests) getUser401(t *testing.T) {
	t.Log("Given the need to validate regular users can't fetch other users.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen fetching the admin user as a regular data.", testID)
		{
			const adminID = "5cf37266-3473-4006-984f-9325122678b7"
			r := httptest.NewRequest(http.MethodGet, "/v1/users/"+adminID, nil)
			w := httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+ut.userToken)
			ut.app.ServeHTTP(w, r)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 401 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 401 for the response.", dbtest.Success, testID)

			recv := w.Body.String()
			resp := `{"error":"Unauthorized"}`
			if resp != recv {
				t.Log("Got :", recv)
				t.Log("Want:", resp)
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)
		}

		testID = 1
		t.Logf("\tTest %d:\tWhen fetching the user as themselves.", testID)
		{
			const userID = "45b5fbd3-755f-4379-8f07-a58d4a30fa2f"
			r := httptest.NewRequest(http.MethodGet, "/v1/users/"+userID, nil)
			w := httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+ut.userToken)
			ut.app.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", dbtest.Success, testID)
		}
	}
}

// getUser404 validates a user request for a user that does not exist with the endpoint.
func (ut *UserTests) getUser404(t *testing.T) {
	id := "c50a5d66-3c4d-453f-af3f-bc960ed1a503"

	r := httptest.NewRequest(http.MethodGet, "/v1/users/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a user with an unknown id.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new user %s.", testID, id)
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 404 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 404 for the response.", dbtest.Success, testID)

			got := w.Body.String()
			exp := "not found"
			if !strings.Contains(got, exp) {
				t.Logf("\t\tTest %d:\tGot : %v", testID, got)
				t.Logf("\t\tTest %d:\tExp: %v", testID, exp)
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)
		}
	}
}

// deleteUserNotFound validates deleting a user that does not exist is not a failure.
func (ut *UserTests) deleteUserNotFound(t *testing.T) {
	id := "a71f77b2-b1ae-4964-a847-f9eecba09d74"

	r := httptest.NewRequest(http.MethodDelete, "/v1/users/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate deleting a user that does not exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new user %s.", testID, id)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", dbtest.Success, testID)
		}
	}
}

// putUser404 validates updating a user that does not exist.
func (ut *UserTests) putUser404(t *testing.T) {
	id := "3097c45e-780a-421b-9eae-43c2fda2bf14"

	u := usergrp.AppUpdateUser{
		Name: dbtest.StringPointer("Doesn't Exist"),
	}
	body, err := json.Marshal(&u)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPut, "/v1/users/"+id, bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate updating a user that does not exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new user %s.", testID, id)
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 404 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 404 for the response.", dbtest.Success, testID)

			got := w.Body.String()
			exp := "not found"
			if !strings.Contains(got, exp) {
				t.Logf("\t\tTest %d:\tGot : %v", testID, got)
				t.Logf("\t\tTest %d:\tExp: %v", testID, exp)
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)
		}
	}
}

// crudUser performs a complete test of CRUD against the api.
func (ut *UserTests) crudUser(t *testing.T) {
	nu := ut.postUser201(t)
	defer ut.deleteUser204(t, nu.ID)

	ut.postUser409(t, nu)

	ut.getUser200(t, nu.ID)
	ut.putUser200(t, nu.ID)
	ut.putUser401(t, nu.ID)
}

// postUser201 validates a user can be created with the endpoint.
func (ut *UserTests) postUser201(t *testing.T) user.User {
	nu := usergrp.AppNewUser{
		Name:            "Bill Kennedy",
		Email:           "bill@ardanlabs.com",
		Roles:           []string{user.RoleAdmin.Name()},
		Password:        "gophers",
		PasswordConfirm: "gophers",
	}

	body, err := json.Marshal(&nu)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)
	ut.app.ServeHTTP(w, r)

	// This needs to be returned for other dbtest.
	var got user.User

	t.Log("Given the need to create a new user with the users endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the declared user value.", testID)
		{
			if w.Code != http.StatusCreated {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 201 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 201 for the response.", dbtest.Success, testID)

			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", dbtest.Failed, testID, err)
			}

			email, err := mail.ParseAddress("bill@ardanlabs.com")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to parse email: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to parse email.", dbtest.Success, testID)

			// Define what we wanted to receive. We will just trust the generated
			// fields like ID and Dates so we copy u.
			exp := got
			exp.Name = "Bill Kennedy"
			exp.Email = *email
			exp.Roles = []user.Role{user.RoleAdmin}

			if diff := cmp.Diff(got, exp); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)
		}
	}

	return got
}

// postUser409 validates a user email field is unique.
func (ut *UserTests) postUser409(t *testing.T, usr user.User) {
	roles := make([]string, len(usr.Roles))
	for i, role := range usr.Roles {
		roles[i] = role.Name()
	}

	nu := usergrp.AppNewUser{
		Name:            usr.Name,
		Email:           usr.Email.Address,
		Roles:           roles,
		Password:        "gophers",
		PasswordConfirm: "gophers",
	}

	body, err := json.Marshal(&nu)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to create a new user with a unique email with the users endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the same declared user value.", testID)
		{
			if w.Code != http.StatusConflict {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 409 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 409 for the response.", dbtest.Success, testID)

		}
	}
}

// deleteUser204 validates deleting a user that does exist.
func (ut *UserTests) deleteUser204(t *testing.T, id uuid.UUID) {
	r := httptest.NewRequest(http.MethodDelete, "/v1/users/"+id.String(), nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate deleting a user that does exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new user %s.", testID, id)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", dbtest.Success, testID)
		}
	}
}

// getUser200 validates a user request for an existing userid.
func (ut *UserTests) getUser200(t *testing.T, id uuid.UUID) {
	r := httptest.NewRequest(http.MethodGet, "/v1/users/"+id.String(), nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a user that exists.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new user %s.", testID, id)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", dbtest.Success, testID)

			var got struct {
				ID           string    `json:"id"`
				Name         string    `json:"name"`
				Email        string    `json:"email"`
				Roles        []string  `json:"roles"`
				PasswordHash []byte    `json:"-"`
				Department   string    `json:"department"`
				Enabled      bool      `json:"enabled"`
				DateCreated  time.Time `json:"dateCreated"`
				DateUpdated  time.Time `json:"dateUpdated"`
			}
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", dbtest.Failed, testID, err)
			}

			email, err := mail.ParseAddress("bill@ardanlabs.com")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to parse email: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to parse email.", dbtest.Success, testID)

			// Define what we wanted to receive. We will just trust the generated
			// fields like Dates so we copy p.
			exp := got
			exp.ID = id.String()
			exp.Name = "Bill Kennedy"
			exp.Email = email.Address
			exp.Roles = []string{user.RoleAdmin.Name()}

			if diff := cmp.Diff(got, exp); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)
		}
	}
}

// putUser200 validates updating a user that does exist.
func (ut *UserTests) putUser200(t *testing.T, id uuid.UUID) {
	u := usergrp.AppUpdateUser{
		Name: dbtest.StringPointer("Bill Kennedy"),
	}
	body, err := json.Marshal(&u)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPut, "/v1/users/"+id.String(), bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to update a user with the users endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the modified user value.", testID)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", dbtest.Success, testID)

			r = httptest.NewRequest(http.MethodGet, "/v1/users/"+id.String(), nil)
			w = httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+ut.adminToken)
			ut.app.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the retrieve : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the retrieve.", dbtest.Success, testID)

			var ru usergrp.AppUser
			if err := json.NewDecoder(w.Body).Decode(&ru); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", dbtest.Failed, testID, err)
			}

			if ru.Name != "Bill Kennedy" {
				t.Fatalf("\t%s\tTest %d:\tShould see an updated Name : got %q want %q", dbtest.Failed, testID, ru.Name, "Bill Kennedy")
			}
			t.Logf("\t%s\tTest %d:\tShould see an updated Name.", dbtest.Success, testID)

			email, err := mail.ParseAddress("bill@ardanlabs.com")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to parse email: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to parse email.", dbtest.Success, testID)

			if ru.Email != email.Address {
				t.Fatalf("\t%s\tTest %d:\tShould not affect other fields like Email : got %q want %q", dbtest.Failed, testID, ru.Email, "bill@ardanlabs.com")
			}
			t.Logf("\t%s\tTest %d:\tShould not affect other fields like Email.", dbtest.Success, testID)
		}
	}
}

// putUser401 validates that a user can't modify users unless they are an admin.
func (ut *UserTests) putUser401(t *testing.T, id uuid.UUID) {
	u := usergrp.AppUpdateUser{
		Name: dbtest.StringPointer("Ale Kennedy"),
	}
	body, err := json.Marshal(&u)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPut, "/v1/users/"+id.String(), bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.userToken)
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to update a user with the users endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen a non-admin user makes a request", testID)
		{
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 401 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 401 for the response.", dbtest.Success, testID)
		}
	}
}

// getUsers200 validates a query request.
func (ut *UserTests) getUsers200(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/users?page=1&rows=2", nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting users.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen querying users", testID)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", dbtest.Success, testID)

			var pr paging.Response[usergrp.AppUser]
			if err := json.Unmarshal(w.Body.Bytes(), &pr); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %s", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to unmarshal the response.", dbtest.Success, testID)

			if pr.Total != 2 || pr.Total != len(pr.Items) {
				t.Log("tot:", pr.Total)
				t.Log("len:", len(pr.Items))
				t.Log("exp:", 2)
				t.Fatalf("\t%s\tTest %d:\tShould get the right number of users.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the right number of users.", dbtest.Success, testID)
		}
	}
}
