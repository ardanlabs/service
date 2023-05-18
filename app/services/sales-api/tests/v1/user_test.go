package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"os"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/ardanlabs/service/app/services/sales-api/handlers"
	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/usergrp"
	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/data/order"
	"github.com/ardanlabs/service/business/sys/validate"
	v1 "github.com/ardanlabs/service/business/web/v1"
	"github.com/ardanlabs/service/business/web/v1/paging"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// UserTests holds methods for each user subtest. This type allows passing
// dependencies for tests while still providing a convenient syntax when
// subtests are registered.
type UserTests struct {
	app        http.Handler
	userToken  string
	adminToken string
}

// Test_Users is the entry point for testing user management apis.
func Test_Users(t *testing.T) {
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

	t.Run("getToken404", tests.getToken404())
	t.Run("getToken200", tests.getToken200())
	t.Run("postUser400", tests.postUser400())
	t.Run("postUser401", tests.postUser401())
	t.Run("postNoAuth401", tests.postNoAuth401())
	t.Run("getUser400", tests.getUser400())
	t.Run("getUser401", tests.getUser401(usrs))
	t.Run("getUser404", tests.getUser404())
	t.Run("deleteUserNotFound", tests.deleteUserNotFound())
	t.Run("putUser404", tests.putUser404())
	t.Run("getUsers200", tests.getUsers200(usrs))
	t.Run("summary", tests.summary(usrs))
	t.Run("crudUsers", tests.crudUser())
}

func (ut *UserTests) getToken404() func(t *testing.T) {
	return func(t *testing.T) {
		url := "/v1/users/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()

		r.SetBasicAuth("unknown@example.com", "some-password")
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Fatalf("Should receive a status code of 404 for the response : %d", w.Code)
		}
	}
}

func (ut *UserTests) getToken200() func(t *testing.T) {
	return func(t *testing.T) {
		url := "/v1/users/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()

		r.SetBasicAuth("admin@example.com", "gophers")
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("Should receive a status code of 200 for the response : %d", w.Code)
		}

		var got struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("Should be able to unmarshal the response : %s", err)
		}
	}
}

func (ut *UserTests) postUser400() func(t *testing.T) {
	return func(t *testing.T) {
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

		if w.Code != http.StatusBadRequest {
			t.Fatalf("Should receive a status code of 400 for the response : %d", w.Code)
		}

		var got v1.ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("Should be able to unmarshal the response to an error type : %s", err)
		}

		fields := validate.FieldErrors{
			{Field: "name", Err: "name is a required field"},
			{Field: "roles", Err: "roles is a required field"},
			{Field: "password", Err: "password is a required field"},
		}
		exp := v1.ErrorResponse{
			Error:  "data validation error",
			Fields: fields.Fields(),
		}

		// We can't rely on the order of the field errors so they have to be
		// sorted. Tell the cmp package how to sort them.
		sorter := cmpopts.SortSlices(func(a, b validate.FieldError) bool {
			return a.Field < b.Field
		})

		if diff := cmp.Diff(got, exp, sorter); diff != "" {
			t.Fatalf("Should get the expected result, diff:\n%s", diff)
		}
	}
}

func (ut *UserTests) postUser401() func(t *testing.T) {
	return func(t *testing.T) {
		body, err := json.Marshal(&usergrp.AppNewUser{})
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.userToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Should receive a status code of 401 for the response : %d", w.Code)
		}
	}
}

func (ut *UserTests) postNoAuth401() func(t *testing.T) {
	return func(t *testing.T) {
		body, err := json.Marshal(&usergrp.AppNewUser{})
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Should receive a status code of 401 for the response : %d", w.Code)
		}
	}
}

func (ut *UserTests) getUser400() func(t *testing.T) {
	return func(t *testing.T) {
		url := fmt.Sprintf("/v1/users/%d", 12345)

		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.adminToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("Should receive a status code of 400 for the response : %d", w.Code)
		}

		got := w.Body.String()
		exp := `{"error":"ID is not in its proper form"}`
		if got != exp {
			t.Logf("got: %v", got)
			t.Logf("exp: %v", exp)
			t.Errorf("Should get the expected result")
		}
	}
}

func (ut *UserTests) getUser401(usrs []user.User) func(t *testing.T) {
	return func(t *testing.T) {
		url := fmt.Sprintf("/v1/users/%s", usrs[0].ID)

		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.userToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Should receive a status code of 401 for the response : %d", w.Code)
		}

		recv := w.Body.String()
		resp := `{"error":"Unauthorized"}`
		if resp != recv {
			t.Log("got:", recv)
			t.Log("exp:", resp)
			t.Fatalf("Should get the expected result.")
		}

		// ---------------------------------------------------------------------

		url = fmt.Sprintf("/v1/users/%s", usrs[1].ID)

		r = httptest.NewRequest(http.MethodGet, url, nil)
		w = httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.userToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("Should receive a status code of 200 for the response : %d", w.Code)
		}
	}
}

func (ut *UserTests) getUser404() func(t *testing.T) {
	return func(t *testing.T) {
		url := fmt.Sprintf("/v1/users/%s", "c50a5d66-3c4d-453f-af3f-bc960ed1a503")

		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.adminToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Fatalf("Should receive a status code of 404 for the response : %d", w.Code)
		}

		got := w.Body.String()
		exp := "not found"
		if !strings.Contains(got, exp) {
			t.Logf("got: %v", got)
			t.Logf("exp: %v", exp)
			t.Errorf("Should get the expected result")
		}
	}
}

func (ut *UserTests) deleteUserNotFound() func(t *testing.T) {
	return func(t *testing.T) {
		url := fmt.Sprintf("/v1/users/%s", "a71f77b2-b1ae-4964-a847-f9eecba09d74")

		r := httptest.NewRequest(http.MethodDelete, url, nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.adminToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusNoContent {
			t.Fatalf("Should receive a status code of 204 for the response : %d", w.Code)
		}
	}
}

func (ut *UserTests) putUser404() func(t *testing.T) {
	return func(t *testing.T) {
		url := fmt.Sprintf("/v1/users/%s", "3097c45e-780a-421b-9eae-43c2fda2bf14")

		u := usergrp.AppUpdateUser{
			Name: dbtest.StringPointer("Doesn't Exist"),
		}
		body, err := json.Marshal(&u)
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.adminToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Fatalf("Should receive a status code of 404 for the response : %d", w.Code)
		}

		got := w.Body.String()
		exp := "not found"
		if !strings.Contains(got, exp) {
			t.Logf("got: %v", got)
			t.Logf("exp: %v", exp)
			t.Errorf("Should get the expected result")
		}
	}
}

func (ut *UserTests) getUsers200(usrs []user.User) func(t *testing.T) {
	return func(t *testing.T) {
		url := "/v1/users?page=1&rows=2"

		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.adminToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("Should receive a status code of 200 for the response : %d", w.Code)
		}

		var pr paging.Response[usergrp.AppUser]
		if err := json.Unmarshal(w.Body.Bytes(), &pr); err != nil {
			t.Fatalf("Should be able to unmarshal the response : %s", err)
		}

		if pr.Total != len(usrs) {
			t.Log("got:", pr.Total)
			t.Log("exp:", len(usrs))
			t.Error("Should get the right total")
		}

		if len(pr.Items) != 2 {
			t.Log("got:", len(pr.Items))
			t.Log("exp:", 2)
			t.Error("Should get the right number of users")
		}
	}
}

func (ut *UserTests) summary(usrs []user.User) func(t *testing.T) {
	return func(t *testing.T) {
		url := "/v1/users/summary"

		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.adminToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("Should receive a status code of 200 for the response : %d", w.Code)
		}

		var pr paging.Response[usergrp.AppSummary]
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

func (ut *UserTests) crudUser() func(t *testing.T) {
	return func(t *testing.T) {
		usr := ut.postUser201(t)
		defer ut.deleteUser204(t, usr.ID)

		ut.postUser409(t, usr)

		ut.getUser200(t, usr.ID)
		ut.putUser200(t, usr.ID)
		ut.putUser401(t, usr.ID)
	}
}

func (ut *UserTests) postUser201(t *testing.T) usergrp.AppUser {
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

	if w.Code != http.StatusCreated {
		t.Fatalf("Should receive a status code of 201 for the response : %d", w.Code)
	}

	var newUsr usergrp.AppUser
	if err := json.NewDecoder(w.Body).Decode(&newUsr); err != nil {
		t.Fatalf("Should be able to unmarshal the response : %s", err)
	}

	email, err := mail.ParseAddress("bill@ardanlabs.com")
	if err != nil {
		t.Fatalf("Should be able to parse email : %s", err)
	}

	exp := newUsr
	exp.Name = "Bill Kennedy"
	exp.Email = email.Address
	exp.Roles = []string{user.RoleAdmin.Name()}

	if diff := cmp.Diff(newUsr, exp); diff != "" {
		t.Fatalf("Should get the expected result, diff:\n%s", diff)
	}

	return newUsr
}

func (ut *UserTests) deleteUser204(t *testing.T, id string) {
	url := fmt.Sprintf("/v1/users/%s", id)

	r := httptest.NewRequest(http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)
	ut.app.ServeHTTP(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Should receive a status code of 204 for the response : %d", w.Code)
	}
}

func (ut *UserTests) postUser409(t *testing.T, usr usergrp.AppUser) {
	nu := usergrp.AppNewUser{
		Name:            usr.Name,
		Email:           usr.Email,
		Roles:           usr.Roles,
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

	if w.Code != http.StatusConflict {
		t.Fatalf("Should receive a status code of 409 for the response : %d", w.Code)
	}
}

func (ut *UserTests) getUser200(t *testing.T, id string) {
	url := fmt.Sprintf("/v1/users/%s", id)

	r := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)
	ut.app.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("Should receive a status code of 200 for the response : %d", w.Code)
	}

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
		t.Fatalf("Should be able to unmarshal the response : %s", err)
	}

	email, err := mail.ParseAddress("bill@ardanlabs.com")
	if err != nil {
		t.Fatalf("Should be able to parse email : %s", err)
	}

	exp := got
	exp.ID = id
	exp.Name = "Bill Kennedy"
	exp.Email = email.Address
	exp.Roles = []string{user.RoleAdmin.Name()}

	if diff := cmp.Diff(got, exp); diff != "" {
		t.Errorf("Should get the expected result, Diff:\n%s", diff)
	}
}

func (ut *UserTests) putUser200(t *testing.T, id string) {
	u := usergrp.AppUpdateUser{
		Name: dbtest.StringPointer("Bill Kennedy"),
	}
	body, err := json.Marshal(&u)
	if err != nil {
		t.Fatal(err)
	}

	url := fmt.Sprintf("/v1/users/%s", id)

	r := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)
	ut.app.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("Should receive a status code of 200 for the response : %d", w.Code)
	}

	r = httptest.NewRequest(http.MethodGet, "/v1/users/"+id, nil)
	w = httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)
	ut.app.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("Should receive a status code of 200 for the retrieve : %d", w.Code)
	}

	var ru usergrp.AppUser
	if err := json.NewDecoder(w.Body).Decode(&ru); err != nil {
		t.Fatalf("Should be able to unmarshal the response : %s", err)
	}

	if ru.Name != "Bill Kennedy" {
		t.Fatalf("Should see an updated Name : got %q want %q", ru.Name, "Bill Kennedy")
	}

	email, err := mail.ParseAddress("bill@ardanlabs.com")
	if err != nil {
		t.Fatalf("Should be able to parse email : %s", err)
	}

	if ru.Email != email.Address {
		t.Fatalf("Should not affect other fields like Email : got %q want %q", ru.Email, "bill@ardanlabs.com")
	}
}

func (ut *UserTests) putUser401(t *testing.T, id string) {
	u := usergrp.AppUpdateUser{
		Name: dbtest.StringPointer("Ale Kennedy"),
	}
	body, err := json.Marshal(&u)
	if err != nil {
		t.Fatal(err)
	}

	url := fmt.Sprintf("/v1/users/%s", id)

	r := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.userToken)
	ut.app.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Should receive a status code of 401 for the response : %d", w.Code)
	}
}
